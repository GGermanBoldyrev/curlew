package ui

import (
	"log/slog"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/GGermanBoldyrev/curlew/internal/collection"
	"github.com/GGermanBoldyrev/curlew/internal/db"
	"github.com/GGermanBoldyrev/curlew/internal/i18n"
	"github.com/GGermanBoldyrev/curlew/internal/paths"
	"github.com/GGermanBoldyrev/curlew/internal/settings"
)

type memoryStore map[string]string

func (m memoryStore) StringWithFallback(key, fallback string) string {
	if v, ok := m[key]; ok {
		return v
	}

	return fallback
}

func (m memoryStore) SetString(key, value string) { m[key] = value }

func openTestStore(t *testing.T) *collection.Store {
	t.Helper()

	handle, err := db.Open(filepath.Join(t.TempDir(), "curlew.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}

	t.Cleanup(func() { handle.Close() })

	return collection.NewStore(handle)
}

func testDeps(t *testing.T) Deps {
	t.Helper()

	catalog, err := i18n.Load()
	if err != nil {
		t.Fatalf("i18n.Load: %v", err)
	}

	translator := catalog.Translator(i18n.Fallback)

	option := settings.Option{
		Key:     settings.KeyLanguage,
		Default: "en",
		Values:  []settings.Value{"en", "ru"},
	}

	return Deps{
		Dirs:        paths.Dirs{Config: t.TempDir(), Cache: t.TempDir()},
		Log:         slog.New(slog.DiscardHandler),
		Settings:    settings.New(memoryStore{}, option),
		T:           translator,
		Collections: collection.New(),
		Store:       openTestStore(t),
	}
}

func TestRestoreSizeFallsBackOnMissingAndAbsurdValues(t *testing.T) {
	tests := []struct {
		name   string
		stored *fyne.Size
		want   fyne.Size
	}{
		{"nothing stored", nil, fyne.NewSize(defaultWindowWidth, defaultWindowHeight)},
		{"sane values survive", sizePtr(1000, 700), fyne.NewSize(1000, 700)},
		{"too narrow falls back", sizePtr(120, 700), fyne.NewSize(defaultWindowWidth, 700)},
		{"too short falls back", sizePtr(1000, 50), fyne.NewSize(1000, defaultWindowHeight)},
		{"absurdly wide falls back", sizePtr(999999, 700), fyne.NewSize(defaultWindowWidth, 700)},
		{"zero falls back", sizePtr(0, 0), fyne.NewSize(defaultWindowWidth, defaultWindowHeight)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefs := test.NewTempApp(t).Preferences()

			if tt.stored != nil {
				saveSize(prefs, *tt.stored)
			}

			if got := restoreSize(prefs); got != tt.want {
				t.Errorf("restoreSize = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSizeRoundTripsThroughPreferences(t *testing.T) {
	prefs := test.NewTempApp(t).Preferences()
	want := fyne.NewSize(1440, 900)

	saveSize(prefs, want)

	if got := restoreSize(prefs); got != want {
		t.Errorf("round trip = %v, want %v", got, want)
	}
}

func TestClosingTheWindowStoresItsSize(t *testing.T) {
	app := test.NewTempApp(t)
	u := New(app, testDeps(t))

	u.win.Resize(fyne.NewSize(1024, 640))
	u.win.Close()

	if got := restoreSize(app.Preferences()); got != fyne.NewSize(1024, 640) {
		t.Errorf("stored size = %v, want 1024x640", got)
	}
}

func sizePtr(w, h float32) *fyne.Size {
	s := fyne.NewSize(w, h)

	return &s
}
