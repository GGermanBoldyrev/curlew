package theme

import (
	"image/color"
	"testing"

	fynetheme "fyne.io/fyne/v2/theme"
)

func TestParseHex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		want    color.NRGBA
		wantErr bool
	}{
		{"six digits get an opaque alpha", "#282C34", color.NRGBA{R: 0x28, G: 0x2C, B: 0x34, A: 0xFF}, false},
		{"eight digits keep their alpha", "#282C3480", color.NRGBA{R: 0x28, G: 0x2C, B: 0x34, A: 0x80}, false},
		{"the hash is optional", "ABB2BF", color.NRGBA{R: 0xAB, G: 0xB2, B: 0xBF, A: 0xFF}, false},
		{"wrong length is rejected", "#FFF", color.NRGBA{}, true},
		{"non hexadecimal is rejected", "#GGGGGG", color.NRGBA{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseHex(tt.in)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseHex(%q) = %v, want an error", tt.in, got)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseHex(%q): %v", tt.in, err)
			}

			if got != tt.want {
				t.Errorf("parseHex(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseVariant(t *testing.T) {
	t.Parallel()

	if got, err := parseVariant("light"); err != nil || got != fynetheme.VariantLight {
		t.Errorf("light = %v, %v", got, err)
	}

	if got, err := parseVariant("dark"); err != nil || got != fynetheme.VariantDark {
		t.Errorf("dark = %v, %v", got, err)
	}

	if _, err := parseVariant("sepia"); err == nil {
		t.Error("parseVariant accepted an unknown variant")
	}
}

func TestPaletteRejectsAnUnknownColorName(t *testing.T) {
	t.Parallel()

	doc := document{Variant: "dark", Colors: map[string]string{"notAColorName": "#000000"}}

	if _, err := doc.palette(knownColors()); err == nil {
		t.Error("an unknown color name was accepted")
	}
}

func TestPaletteRejectsAMalformedColor(t *testing.T) {
	t.Parallel()

	doc := document{Variant: "dark", Colors: map[string]string{"background": "not-a-color"}}

	if _, err := doc.palette(knownColors()); err == nil {
		t.Error("a malformed color was accepted")
	}
}
