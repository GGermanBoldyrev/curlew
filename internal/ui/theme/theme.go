package theme

import (
	"embed"
	"encoding/json"
	"fmt"
	"image/color"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"
)

const Default = "dark"

//go:embed themes/*.json
var files embed.FS

type Palette struct {
	Variant fyne.ThemeVariant
	Colors  map[fyne.ThemeColorName]color.Color
}

type Named struct {
	Name    string
	Palette Palette
}

type Set struct {
	themes []Named
}

func Load() (*Set, error) {
	const dir = "themes"

	entries, err := files.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}

	known := knownColors()
	themes := make([]Named, 0, len(entries))

	for _, entry := range entries {
		theme, err := decode(path.Join(dir, entry.Name()), known)
		if err != nil {
			return nil, err
		}

		themes = append(themes, theme)
	}

	return &Set{themes: themes}, nil
}

func decode(file string, known map[string]fyne.ThemeColorName) (Named, error) {
	name := path.Base(file)

	raw, readErr := fs.ReadFile(files, file)
	if readErr != nil {
		return Named{}, fmt.Errorf("read %s: %w", name, readErr)
	}

	var doc document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return Named{}, fmt.Errorf("parse %s: %w", name, err)
	}

	palette, err := doc.palette(known)
	if err != nil {
		return Named{}, fmt.Errorf("%s: %w", name, err)
	}

	return Named{Name: strings.TrimSuffix(name, ".json"), Palette: palette}, nil
}

func (s *Set) Names() []string {
	names := make([]string, 0, len(s.themes))
	for _, t := range s.themes {
		names = append(names, t.Name)
	}

	return names
}

func (s *Set) ByName(name string) (Palette, bool) {
	i := slices.IndexFunc(s.themes, func(t Named) bool { return t.Name == name })
	if i < 0 {
		return Palette{}, false
	}

	return s.themes[i].Palette, true
}

func New(p Palette) fyne.Theme {
	return themed{Theme: fynetheme.DefaultTheme(), colors: p.Colors, variant: p.Variant}
}

type themed struct {
	fyne.Theme

	colors  map[fyne.ThemeColorName]color.Color
	variant fyne.ThemeVariant
}

func (t themed) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if c, ok := t.colors[name]; ok {
		return c
	}

	return t.Theme.Color(name, t.variant)
}

type document struct {
	Variant string            `json:"variant"`
	Colors  map[string]string `json:"colors"`
}

func (d document) palette(known map[string]fyne.ThemeColorName) (Palette, error) {
	variant, err := parseVariant(d.Variant)
	if err != nil {
		return Palette{}, err
	}

	colors := make(map[fyne.ThemeColorName]color.Color, len(d.Colors))

	for key, value := range d.Colors {
		name, ok := known[key]
		if !ok {
			return Palette{}, fmt.Errorf("unknown color %q", key)
		}

		parsed, err := parseHex(value)
		if err != nil {
			return Palette{}, fmt.Errorf("color %q: %w", key, err)
		}

		colors[name] = parsed
	}

	return Palette{Variant: variant, Colors: colors}, nil
}

func parseVariant(name string) (fyne.ThemeVariant, error) {
	switch name {
	case "light":
		return fynetheme.VariantLight, nil
	case "dark":
		return fynetheme.VariantDark, nil
	default:
		return 0, fmt.Errorf("unknown variant %q, want light or dark", name)
	}
}

func parseHex(value string) (color.NRGBA, error) {
	digits := strings.TrimPrefix(value, "#")
	if len(digits) == 6 {
		digits += "FF"
	}

	if len(digits) != 8 {
		return color.NRGBA{}, fmt.Errorf("want #RRGGBB or #RRGGBBAA, got %q", value)
	}

	var channels [4]uint8

	for i := range channels {
		parsed, err := strconv.ParseUint(digits[i*2:i*2+2], 16, 8)
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("not hexadecimal: %q", value)
		}

		channels[i] = uint8(parsed)
	}

	return color.NRGBA{R: channels[0], G: channels[1], B: channels[2], A: channels[3]}, nil
}

func knownColors() map[string]fyne.ThemeColorName {
	names := []fyne.ThemeColorName{
		fynetheme.ColorNameBackground,
		fynetheme.ColorNameButton,
		fynetheme.ColorNameDisabled,
		fynetheme.ColorNameDisabledButton,
		fynetheme.ColorNameError,
		fynetheme.ColorNameFocus,
		fynetheme.ColorNameForeground,
		fynetheme.ColorNameForegroundOnError,
		fynetheme.ColorNameForegroundOnPrimary,
		fynetheme.ColorNameForegroundOnSuccess,
		fynetheme.ColorNameForegroundOnWarning,
		fynetheme.ColorNameHeaderBackground,
		fynetheme.ColorNameHover,
		fynetheme.ColorNameHyperlink,
		fynetheme.ColorNameInnerWindowBorder,
		fynetheme.ColorNameInnerWindowBorderInactive,
		fynetheme.ColorNameInputBackground,
		fynetheme.ColorNameInputBorder,
		fynetheme.ColorNameMenuBackground,
		fynetheme.ColorNameOverlayBackground,
		fynetheme.ColorNamePlaceHolder,
		fynetheme.ColorNamePressed,
		fynetheme.ColorNamePrimary,
		fynetheme.ColorNameScrollBar,
		fynetheme.ColorNameScrollBarBackground,
		fynetheme.ColorNameSelection,
		fynetheme.ColorNameSeparator,
		fynetheme.ColorNameShadow,
		fynetheme.ColorNameSuccess,
		fynetheme.ColorNameWarning,
	}

	known := make(map[string]fyne.ThemeColorName, len(names))
	for _, name := range names {
		known[string(name)] = name
	}

	return known
}
