package theme_test

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/test"
	fynetheme "fyne.io/fyne/v2/theme"

	"github.com/GGermanBoldyrev/curlew/internal/ui/theme"
)

func load(t *testing.T) *theme.Set {
	t.Helper()

	set, err := theme.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	return set
}

func TestEveryEmbeddedThemeParses(t *testing.T) {
	t.Parallel()

	set := load(t)

	if len(set.Names()) < 2 {
		t.Fatalf("names = %v, want at least two themes", set.Names())
	}
}

func TestNamesAreUniqueAndContainTheDefault(t *testing.T) {
	t.Parallel()

	set := load(t)

	seen := make(map[string]bool)
	for _, name := range set.Names() {
		if seen[name] {
			t.Errorf("duplicate theme name %q", name)
		}

		seen[name] = true
	}

	if !seen[theme.Default] {
		t.Errorf("default %q is not among %v", theme.Default, set.Names())
	}
}

func TestByNameResolvesEveryTheme(t *testing.T) {
	t.Parallel()

	set := load(t)

	for _, name := range set.Names() {
		if _, ok := set.ByName(name); !ok {
			t.Errorf("ByName(%q) not found", name)
		}
	}

	if _, ok := set.ByName("no-such-theme"); ok {
		t.Error("ByName accepted an unknown name")
	}
}

func TestPaletteColorsWinOverTheRequestedVariant(t *testing.T) {
	test.NewTempApp(t)

	set := load(t)

	palette, ok := set.ByName("onedark")
	if !ok {
		t.Fatal("onedark is not registered")
	}

	got := theme.New(palette).Color(fynetheme.ColorNameBackground, fynetheme.VariantLight)
	want := color.NRGBA{R: 0x28, G: 0x2C, B: 0x34, A: 0xFF}

	if !sameColor(got, want) {
		t.Errorf("background = %v, want the One Dark background %v", got, want)
	}
}

func TestThemeWithoutOverridesFollowsItsOwnVariant(t *testing.T) {
	test.NewTempApp(t)

	set := load(t)

	dark, ok := set.ByName("dark")
	if !ok {
		t.Fatal("dark is not registered")
	}

	light, ok := set.ByName("light")
	if !ok {
		t.Fatal("light is not registered")
	}

	darkBg := theme.New(dark).Color(fynetheme.ColorNameBackground, fynetheme.VariantLight)
	lightBg := theme.New(light).Color(fynetheme.ColorNameBackground, fynetheme.VariantDark)

	if sameColor(darkBg, lightBg) {
		t.Error("dark and light resolved to the same background, the variant is being ignored")
	}
}

func sameColor(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()

	return ar == br && ag == bg && ab == bb && aa == ba
}
