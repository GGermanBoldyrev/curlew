package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	"github.com/GGermanBoldyrev/curlew/internal/collection"
	"github.com/GGermanBoldyrev/curlew/internal/db"
	"github.com/GGermanBoldyrev/curlew/internal/i18n"
	"github.com/GGermanBoldyrev/curlew/internal/paths"
	"github.com/GGermanBoldyrev/curlew/internal/settings"
	"github.com/GGermanBoldyrev/curlew/internal/ui"
	"github.com/GGermanBoldyrev/curlew/internal/ui/theme"
)

const appID = "io.github.ggermanboldyrev.curlew"

type Config struct {
	Verbose bool
}

type App struct {
	log  *slog.Logger
	dirs paths.Dirs

	fyne fyne.App
	ui   *ui.UI
	db   *sql.DB
}

func New(cfg Config) (*App, error) {
	level := slog.LevelInfo
	if cfg.Verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	dirs, err := resolveDirs()
	if err != nil {
		return nil, err
	}

	fa := fyneapp.NewWithID(appID)

	themes, err := theme.Load()
	if err != nil {
		return nil, err
	}

	catalog, err := i18n.Load()
	if err != nil {
		return nil, err
	}

	appearance, language := themeOption(themes), languageOption(catalog)
	registry := settings.New(fa.Preferences(), appearance, language)

	translator := catalog.Translator(i18n.Lang(registry.Get(language)))

	applyTheme(fa, themes, registry.Get(appearance))

	a := &App{log: log, dirs: dirs, fyne: fa}
	database, err := db.Open(dirs.DB())
	if err != nil {
		return nil, err
	}

	a.db = database
	store := collection.NewStore(database)

	collections, err := store.Load()
	if err != nil {
		return nil, errors.Join(err, database.Close())
	}

	log.Debug("collections loaded",
		slog.String("database", dirs.DB()), slog.Int("count", collections.Len()))

	a.ui = ui.New(fa, ui.Deps{
		Dirs:        dirs,
		Log:         log,
		Settings:    registry,
		T:           translator,
		Collections: collections,
		Store:       store,
	})

	registry.OnChange(func(o settings.Option, v settings.Value) {
		switch o.Key {
		case settings.KeyTheme:
			applyTheme(fa, themes, v)
		case settings.KeyLanguage:
			translator.SetLang(i18n.Lang(v))
			a.ui.Rebuild()
		}

		log.Debug("setting changed", slog.String("key", o.Key), slog.String("value", string(v)))
	})

	log.Debug("app constructed", slog.String("config_dir", dirs.Config), slog.String("cache_dir", dirs.Cache))
	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			fyne.Do(a.fyne.Quit)
		case <-stopped:
		}
	}()

	a.ui.Window().Show()
	a.fyne.Run()
	close(stopped)

	if err := ctx.Err(); err != nil && !isCancelled(err) {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}

func (a *App) Close() error {
	if a.db == nil {
		return nil
	}

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}

	return nil
}

func resolveDirs() (paths.Dirs, error) {
	dirs, err := paths.Resolve()
	if err != nil {
		return paths.Dirs{}, err
	}

	if err := dirs.MkdirAll(); err != nil {
		return paths.Dirs{}, err
	}

	return dirs, nil
}

func themeOption(themes *theme.Set) settings.Option {
	names := themes.Names()

	values := make([]settings.Value, 0, len(names))
	for _, name := range names {
		values = append(values, settings.Value(name))
	}

	return settings.Option{
		Key:     settings.KeyTheme,
		Default: theme.Default,
		Values:  values,
	}
}

func applyTheme(a fyne.App, themes *theme.Set, name settings.Value) {
	palette, ok := themes.ByName(string(name))
	if !ok {
		return
	}

	a.Settings().SetTheme(theme.New(palette))
}

func languageOption(catalog *i18n.Catalog) settings.Option {
	langs := catalog.Langs()

	values := make([]settings.Value, 0, len(langs))
	for _, lang := range langs {
		values = append(values, settings.Value(lang))
	}

	return settings.Option{
		Key:     settings.KeyLanguage,
		Default: settings.Value(i18n.Fallback),
		Values:  values,
	}
}

func isCancelled(err error) bool { return errors.Is(err, context.Canceled) }
