package app

import (
	"testing"

	"github.com/GGermanBoldyrev/curlew/internal/i18n"
	"github.com/GGermanBoldyrev/curlew/internal/settings"
	"github.com/GGermanBoldyrev/curlew/internal/ui/theme"
)

func TestEveryOptionValueHasAnEnglishLabel(t *testing.T) {
	t.Parallel()

	catalog, err := i18n.Load()
	if err != nil {
		t.Fatalf("i18n.Load: %v", err)
	}

	themes, err := theme.Load()
	if err != nil {
		t.Fatalf("theme.Load: %v", err)
	}

	for _, option := range []settings.Option{themeOption(themes), languageOption(catalog)} {
		if !catalog.Has(i18n.Fallback, option.NameKey()) {
			t.Errorf("no translation for %q", option.NameKey())
		}

		for _, value := range option.Values {
			if !catalog.Has(i18n.Fallback, option.ValueKey(value)) {
				t.Errorf("no translation for %q", option.ValueKey(value))
			}
		}
	}
}
