package collection_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/GGermanBoldyrev/curlew/internal/collection"
	"github.com/GGermanBoldyrev/curlew/internal/db"
)

func openStore(t *testing.T) *collection.Store {
	t.Helper()

	return collection.NewStore(openDB(t, filepath.Join(t.TempDir(), "curlew.db")))
}

func openDB(t *testing.T, path string) *sql.DB {
	t.Helper()

	handle, err := db.Open(path)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}

	t.Cleanup(func() {
		if err := handle.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	return handle
}

func names(t *testing.T, tree *collection.Tree, parent string) []string {
	t.Helper()

	children := tree.Children(parent)

	out := make([]string, 0, len(children))
	for _, id := range children {
		name, _ := tree.Name(id)
		out = append(out, name)
	}

	return out
}

func mustCreate(t *testing.T, store *collection.Store, parent, name string) string {
	t.Helper()

	id, err := store.Create(parent, name)
	if err != nil {
		t.Fatalf("Create(%q, %q): %v", parent, name, err)
	}

	return id
}

func TestFreshDatabaseLoadsAnEmptyTree(t *testing.T) {
	t.Parallel()

	tree, err := openStore(t).Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := tree.Len(); got != 0 {
		t.Errorf("Len = %d, want 0", got)
	}
}

func TestCreateThenLoadRoundTrips(t *testing.T) {
	t.Parallel()

	store := openStore(t)

	api := mustCreate(t, store, collection.Root, "Payments API")
	users := mustCreate(t, store, api, "Users")
	mustCreate(t, store, users, "Create user")
	mustCreate(t, store, collection.Root, "Internal tools")

	tree, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := tree.Len(); got != 4 {
		t.Fatalf("Len = %d, want 4", got)
	}

	if got := names(t, tree, collection.Root); len(got) != 2 || got[0] != "Internal tools" {
		t.Errorf("roots = %v, want them sorted by name", got)
	}

	if got := names(t, tree, api); len(got) != 1 || got[0] != "Users" {
		t.Errorf("children of the API = %v", got)
	}
}

func TestIdentifiersSurviveAReload(t *testing.T) {
	t.Parallel()

	store := openStore(t)
	api := mustCreate(t, store, collection.Root, "API")

	tree, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if _, ok := tree.Name(api); !ok {
		t.Errorf("the identifier %q returned by Create is not in the loaded tree", api)
	}
}

func TestDuplicateSiblingNamesAreAllowed(t *testing.T) {
	t.Parallel()

	store := openStore(t)

	first := mustCreate(t, store, collection.Root, "Users")
	second := mustCreate(t, store, collection.Root, "Users")

	if first == second {
		t.Fatal("two collections were given the same identifier")
	}

	tree, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := names(t, tree, collection.Root); len(got) != 2 {
		t.Errorf("roots = %v, want both duplicates", got)
	}
}

func TestRenameTouchesOnlyTheChosenRow(t *testing.T) {
	t.Parallel()

	store := openStore(t)

	first := mustCreate(t, store, collection.Root, "Users")
	mustCreate(t, store, collection.Root, "Users")

	if err := store.Rename(first, "Members"); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	tree, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if name, _ := tree.Name(first); name != "Members" {
		t.Errorf("renamed = %q", name)
	}

	if got := names(t, tree, collection.Root); len(got) != 2 {
		t.Fatalf("roots = %v", got)
	}
}

func TestRemoveCascadesToDescendants(t *testing.T) {
	t.Parallel()

	store := openStore(t)

	api := mustCreate(t, store, collection.Root, "API")
	users := mustCreate(t, store, api, "Users")
	mustCreate(t, store, users, "Create user")
	mustCreate(t, store, collection.Root, "Internal")

	if err := store.Remove(api); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	tree, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := tree.Len(); got != 1 {
		t.Errorf("Len = %d, want only the untouched sibling", got)
	}
}

func TestOperationsOnUnknownIdentifiers(t *testing.T) {
	t.Parallel()

	store := openStore(t)

	if err := store.Rename("4242", "X"); !errors.Is(err, collection.ErrNotFound) {
		t.Errorf("rename = %v, want ErrNotFound", err)
	}

	if err := store.Remove("4242"); !errors.Is(err, collection.ErrNotFound) {
		t.Errorf("remove = %v, want ErrNotFound", err)
	}

	if err := store.Rename("not-a-number", "X"); !errors.Is(err, collection.ErrNotFound) {
		t.Errorf("malformed identifier = %v, want ErrNotFound", err)
	}

	if _, err := store.Create("4242", "Child"); !errors.Is(err, collection.ErrParentUnknown) {
		t.Errorf("unknown parent = %v, want ErrParentUnknown", err)
	}
}

func TestCreateRejectsABlankName(t *testing.T) {
	t.Parallel()

	if _, err := openStore(t).Create(collection.Root, "   "); !errors.Is(err, collection.ErrEmptyName) {
		t.Errorf("err = %v, want ErrEmptyName", err)
	}
}
