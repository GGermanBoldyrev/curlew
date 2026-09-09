package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/pressly/goose/v3"

	_ "modernc.org/sqlite"
)

const (
	driver        = "sqlite"
	migrationsDir = "migrations"
	timeout       = 30 * time.Second
)

//go:embed migrations/*.sql
var files embed.FS

func Open(path string) (*sql.DB, error) {
	dsn := "file:" + path +
		"?_pragma=foreign_keys(1)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(5000)"

	handle, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	handle.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := Migrate(ctx, handle); err != nil {
		return nil, errors.Join(err, handle.Close())
	}

	return handle, nil
}

func Migrate(ctx context.Context, handle *sql.DB) error {
	sources, err := fs.Sub(files, migrationsDir)
	if err != nil {
		return fmt.Errorf("locate migrations: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, handle, sources,
		goose.WithDisableGlobalRegistry(true))
	if err != nil {
		return fmt.Errorf("build migration provider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
