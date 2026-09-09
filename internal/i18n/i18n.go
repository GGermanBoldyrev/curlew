package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"
	"sync"
)

const Fallback Lang = "en"

//go:embed locales/*.json
var locales embed.FS

type Lang string

type Catalog struct {
	tables map[Lang]map[string]string
}

func Load() (*Catalog, error) {
	const dir = "locales"

	entries, err := locales.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}

	tables := make(map[Lang]map[string]string, len(entries))

	for _, entry := range entries {
		lang, table, err := decode(path.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		tables[lang] = table
	}

	return NewCatalog(tables), nil
}

func NewCatalog(tables map[Lang]map[string]string) *Catalog {
	return &Catalog{tables: tables}
}

func (c *Catalog) Langs() []Lang {
	langs := slices.Collect(maps.Keys(c.tables))
	slices.Sort(langs)

	return langs
}

func (c *Catalog) Keys(lang Lang) []string {
	keys := slices.Collect(maps.Keys(c.tables[lang]))
	slices.Sort(keys)

	return keys
}

func (c *Catalog) Has(lang Lang, key string) bool {
	_, ok := c.tables[lang][key]

	return ok
}

func (c *Catalog) Translator(lang Lang) *Translator {
	return &Translator{catalog: c, lang: lang}
}

type Translator struct {
	mu      sync.RWMutex
	catalog *Catalog
	lang    Lang
}

func (t *Translator) Lang() Lang {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.lang
}

func (t *Translator) SetLang(lang Lang) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lang = lang
}

func (t *Translator) T(key string) string {
	t.mu.RLock()
	lang := t.lang
	t.mu.RUnlock()

	if v, ok := t.catalog.tables[lang][key]; ok {
		return v
	}

	if v, ok := t.catalog.tables[Fallback][key]; ok {
		return v
	}

	return key
}

func decode(file string) (Lang, map[string]string, error) {
	name := path.Base(file)

	raw, readErr := fs.ReadFile(locales, file)
	if readErr != nil {
		return "", nil, fmt.Errorf("read %s: %w", name, readErr)
	}

	table := make(map[string]string)
	if err := json.Unmarshal(raw, &table); err != nil {
		return "", nil, fmt.Errorf("parse %s: %w", name, err)
	}

	return Lang(strings.TrimSuffix(name, ".json")), table, nil
}
