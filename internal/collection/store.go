package collection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const storeTimeout = 5 * time.Second

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Load() (*Tree, error) {
	ctx, cancel := context.WithTimeout(context.Background(), storeTimeout)
	defer cancel()

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, COALESCE(parent_id, 0), name FROM collections ORDER BY name COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("read collections: %w", err)
	}
	defer rows.Close()

	type record struct {
		id, parent int64
		name       string
	}

	byParent := make(map[int64][]record)

	for rows.Next() {
		var r record
		if err := rows.Scan(&r.id, &r.parent, &r.name); err != nil {
			return nil, fmt.Errorf("scan collection: %w", err)
		}

		byParent[r.parent] = append(byParent[r.parent], r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read collections: %w", err)
	}

	tree := New()

	var attach func(int64, string)
	attach = func(dbParent int64, treeParent string) {
		for _, r := range byParent[dbParent] {
			id := strconv.FormatInt(r.id, 10)

			if err := tree.Insert(id, treeParent, r.name); err != nil {
				continue
			}

			attach(r.id, id)
		}
	}

	attach(0, Root)

	return tree, nil
}

func (s *Store) Create(parent, name string) (string, error) {
	name, err := clean(name)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), storeTimeout)
	defer cancel()

	reference, err := identify(parent)
	if err != nil {
		return "", err
	}

	var id int64

	if err := s.tx(ctx, func(tx *sql.Tx) error {
		if missing := exists(ctx, tx, reference); missing != nil {
			return missing
		}

		result, insertErr := tx.ExecContext(ctx,
			`INSERT INTO collections (parent_id, name) VALUES (?, ?)`, nullable(reference), name)
		if insertErr != nil {
			return fmt.Errorf("insert collection: %w", insertErr)
		}

		newID, idErr := result.LastInsertId()
		if idErr != nil {
			return fmt.Errorf("read new identifier: %w", idErr)
		}

		id = newID

		return nil
	}); err != nil {
		return "", err
	}

	return strconv.FormatInt(id, 10), nil
}

func (s *Store) Rename(id, name string) error {
	name, err := clean(name)
	if err != nil {
		return err
	}

	reference, err := identify(id)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), storeTimeout)
	defer cancel()

	return s.tx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE collections SET name = ? WHERE id = ?`, name, reference)
		if err != nil {
			return fmt.Errorf("rename collection: %w", err)
		}

		return touched(result)
	})
}

func (s *Store) Remove(id string) error {
	reference, err := identify(id)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), storeTimeout)
	defer cancel()

	return s.tx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM collections WHERE id = ?`, reference)
		if err != nil {
			return fmt.Errorf("delete collection: %w", err)
		}

		return touched(result)
	})
}

func (s *Store) tx(ctx context.Context, body func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	if err := body(tx); err != nil {
		return errors.Join(err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func exists(ctx context.Context, tx *sql.Tx, reference int64) error {
	if reference == 0 {
		return nil
	}

	var found int64

	err := tx.QueryRowContext(ctx, `SELECT id FROM collections WHERE id = ?`, reference).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrParentUnknown
	}

	if err != nil {
		return fmt.Errorf("look up parent: %w", err)
	}

	return nil
}

func identify(id string) (int64, error) {
	if id == Root {
		return 0, nil
	}

	reference, err := strconv.ParseInt(id, 10, 64)
	if err != nil || reference <= 0 {
		return 0, fmt.Errorf("%w: %q", ErrNotFound, id)
	}

	return reference, nil
}

func touched(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count affected rows: %w", err)
	}

	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func nullable(reference int64) any {
	if reference == 0 {
		return nil
	}

	return reference
}
