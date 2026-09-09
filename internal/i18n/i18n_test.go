package i18n_test

import (
	"testing"

	"github.com/GGermanBoldyrev/curlew/internal/i18n"
)

func loadCatalog(t *testing.T) *i18n.Catalog {
	t.Helper()

	catalog, err := i18n.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	return catalog
}

func TestEveryLanguageHasEveryKey(t *testing.T) {
	t.Parallel()

	catalog := loadCatalog(t)
	base := catalog.Keys(i18n.Fallback)

	if len(base) == 0 {
		t.Fatalf("the %q table is empty", i18n.Fallback)
	}

	for _, lang := range catalog.Langs() {
		for _, key := range base {
			if !catalog.Has(lang, key) {
				t.Errorf("%s.json is missing key %q", lang, key)
			}
		}

		for _, key := range catalog.Keys(lang) {
			if !catalog.Has(i18n.Fallback, key) {
				t.Errorf("%s.json has key %q that %s.json does not", lang, key, i18n.Fallback)
			}
		}
	}
}

func TestLangsAreSortedAndContainTheFallback(t *testing.T) {
	t.Parallel()

	langs := loadCatalog(t).Langs()

	if len(langs) < 2 {
		t.Fatalf("langs = %v, want at least two", langs)
	}

	if langs[0] != i18n.Fallback {
		t.Errorf("langs = %v, want %q first", langs, i18n.Fallback)
	}
}

func TestTranslationFallsBackToEnglishThenToTheKey(t *testing.T) {
	t.Parallel()

	catalog := i18n.NewCatalog(map[i18n.Lang]map[string]string{
		"en": {"translated": "Send", "english.only": "English only"},
		"ru": {"translated": "Отправить"},
	})

	translator := catalog.Translator("ru")

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"active language wins", "translated", "Отправить"},
		{"missing key falls back to english", "english.only", "English only"},
		{"unknown key returns itself", "no.such.key", "no.such.key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := translator.T(tt.key); got != tt.want {
				t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestSetLangSwitchesTable(t *testing.T) {
	t.Parallel()

	catalog := i18n.NewCatalog(map[i18n.Lang]map[string]string{
		"en": {"k": "English"},
		"ru": {"k": "Русский"},
	})

	translator := catalog.Translator("en")

	if got := translator.T("k"); got != "English" {
		t.Fatalf("T = %q", got)
	}

	translator.SetLang("ru")

	if got := translator.T("k"); got != "Русский" {
		t.Errorf("after SetLang T = %q", got)
	}

	if translator.Lang() != "ru" {
		t.Errorf("Lang = %q", translator.Lang())
	}
}

func TestOneCatalogServesIndependentTranslators(t *testing.T) {
	t.Parallel()

	catalog := loadCatalog(t)

	english := catalog.Translator("en")
	russian := catalog.Translator("ru")

	russian.SetLang("ru")

	if english.Lang() != "en" {
		t.Errorf("translators share state, english became %q", english.Lang())
	}
}
