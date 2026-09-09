package ui_test

import (
	"log/slog"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/GGermanBoldyrev/curlew/internal/collection"
	"github.com/GGermanBoldyrev/curlew/internal/db"
	"github.com/GGermanBoldyrev/curlew/internal/i18n"
	"github.com/GGermanBoldyrev/curlew/internal/paths"
	"github.com/GGermanBoldyrev/curlew/internal/settings"
	"github.com/GGermanBoldyrev/curlew/internal/ui"
)

type fakeStore map[string]string

func (f fakeStore) StringWithFallback(key, fallback string) string {
	if v, ok := f[key]; ok {
		return v
	}

	return fallback
}

func (f fakeStore) SetString(key, value string) {
	f[key] = value
}

func language() settings.Option {
	return settings.Option{
		Key:     settings.KeyLanguage,
		Default: "en",
		Values:  []settings.Value{"en", "ru"},
	}
}

func newStore(t *testing.T) *collection.Store {
	t.Helper()

	handle, err := db.Open(filepath.Join(t.TempDir(), "curlew.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}

	t.Cleanup(func() { handle.Close() })

	return collection.NewStore(handle)
}

func newUI(t *testing.T) (*ui.UI, *settings.Registry, *i18n.Translator) {
	t.Helper()

	registry := settings.New(fakeStore{}, language())

	catalog, err := i18n.Load()
	if err != nil {
		t.Fatalf("i18n.Load: %v", err)
	}

	translator := catalog.Translator(i18n.Fallback)

	u := ui.New(test.NewTempApp(t), ui.Deps{
		Dirs:        paths.Dirs{Config: t.TempDir(), Cache: t.TempDir()},
		Log:         slog.New(slog.DiscardHandler),
		Settings:    registry,
		T:           translator,
		Collections: collection.New(),
		Store:       newStore(t),
	})

	return u, registry, translator
}

func findSelect(o fyne.CanvasObject) *widget.Select {
	switch v := o.(type) {
	case *widget.Select:
		return v
	case *fyne.Container:
		for _, child := range v.Objects {
			if found := findSelect(child); found != nil {
				return found
			}
		}
	}

	return nil
}

func TestNewCreatesAnUntitledWindowWithContent(t *testing.T) {
	u, _, _ := newUI(t)

	if u.Window().Content() == nil {
		t.Fatal("window has no content")
	}

	if got := u.Window().Title(); got != "" {
		t.Errorf("title = %q, want it empty", got)
	}
}

func TestHeaderCarriesASelectPerOption(t *testing.T) {
	u, _, _ := newUI(t)

	sel := findSelect(u.Window().Content())
	if sel == nil {
		t.Fatal("no select found in the header")
	}

	if got := sel.Selected; got != "English" {
		t.Errorf("selected = %q, want %q", got, "English")
	}
}

func TestSwitchingLanguageRelabelsTheHeader(t *testing.T) {
	u, registry, translator := newUI(t)

	registry.Set(language(), "ru")
	translator.SetLang("ru")
	u.Rebuild()

	sel := findSelect(u.Window().Content())
	if sel == nil {
		t.Fatal("no select found after rebuild")
	}

	if got := sel.Selected; got != "Русский" {
		t.Errorf("selected = %q, want %q", got, "Русский")
	}

	if got := sel.Options; len(got) != 2 || got[0] != "Английский" {
		t.Errorf("options = %v, want them translated", got)
	}
}
