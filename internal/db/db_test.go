package db_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/GGermanBoldyrev/curlew/internal/db"
)

func TestOpenAppliesTheSchema(t *testing.T) {
	t.Parallel()

	handle, err := db.Open(filepath.Join(t.TempDir(), "curlew.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer handle.Close()

	var name string

	if err := handle.QueryRow(
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'collections'`,
	).Scan(&name); err != nil {
		t.Fatalf("the collections table was not created: %v", err)
	}
}

func TestForeignKeysAreOnSoCascadeWorks(t *testing.T) {
	t.Parallel()

	handle, err := db.Open(filepath.Join(t.TempDir(), "curlew.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer handle.Close()

	var enabled int
	if err := handle.QueryRow(`PRAGMA foreign_keys`).Scan(&enabled); err != nil {
		t.Fatalf("read pragma: %v", err)
	}

	if enabled != 1 {
		t.Error("foreign keys are off, ON DELETE CASCADE would silently do nothing")
	}
}

func TestEveryMigrationIsRecordedByVersion(t *testing.T) {
	t.Parallel()

	handle, err := db.Open(filepath.Join(t.TempDir(), "curlew.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer handle.Close()

	var applied int
	if err := handle.QueryRow(
		`SELECT COUNT(*) FROM goose_db_version WHERE is_applied = 1 AND version_id > 0`,
	).Scan(&applied); err != nil {
		t.Fatalf("read goose bookkeeping: %v", err)
	}

	if applied == 0 {
		t.Fatal("goose recorded no migrations")
	}

	var distinct int
	if err := handle.QueryRow(
		`SELECT COUNT(DISTINCT version_id) FROM goose_db_version WHERE is_applied = 1 AND version_id > 0`,
	).Scan(&distinct); err != nil {
		t.Fatalf("count versions: %v", err)
	}

	if distinct != applied {
		t.Errorf("recorded %d rows for %d distinct versions, a migration ran twice", applied, distinct)
	}
}

func TestReopeningDoesNotReapplyMigrations(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "curlew.db")

	first, openErr := db.Open(path)
	if openErr != nil {
		t.Fatalf("Open: %v", openErr)
	}

	before := appliedCount(t, first)

	if _, err := first.Exec(`INSERT INTO collections (parent_id, name) VALUES (NULL, 'API')`); err != nil {
		t.Fatalf("insert: %v", err)
	}

	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second, reopenErr := db.Open(path)
	if reopenErr != nil {
		t.Fatalf("reopen: %v", reopenErr)
	}
	defer second.Close()

	var rows int
	if err := second.QueryRow(`SELECT COUNT(*) FROM collections`).Scan(&rows); err != nil {
		t.Fatalf("count: %v", err)
	}

	if rows != 1 {
		t.Errorf("rows = %d, a migration re-ran and wiped the data", rows)
	}

	if after := appliedCount(t, second); after != before {
		t.Errorf("applied migrations went from %d to %d, the schema was re-applied", before, after)
	}
}

func TestOpenFailsOnAnUnwritablePath(t *testing.T) {
	t.Parallel()

	if _, err := db.Open(filepath.Join(t.TempDir(), "missing", "curlew.db")); err == nil {
		t.Error("opening below a missing directory should fail")
	}
}

func appliedCount(t *testing.T, handle *sql.DB) int {
	t.Helper()

	var applied int

	if err := handle.QueryRow(
		`SELECT COUNT(*) FROM goose_db_version WHERE is_applied = 1 AND version_id > 0`,
	).Scan(&applied); err != nil {
		t.Fatalf("read goose bookkeeping: %v", err)
	}

	return applied
}
