package collection_test

import (
	"errors"
	"testing"

	"github.com/GGermanBoldyrev/curlew/internal/collection"
)

func TestInsertAtRoot(t *testing.T) {
	t.Parallel()

	tree := collection.New()

	if err := tree.Insert("1", collection.Root, "Payments API"); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	name, ok := tree.Name("1")
	if !ok || name != "Payments API" {
		t.Errorf("Name = %q, %v", name, ok)
	}

	if got := tree.Children(collection.Root); len(got) != 1 || got[0] != "1" {
		t.Errorf("roots = %v", got)
	}

	if got := tree.Depth("1"); got != 1 {
		t.Errorf("depth = %d, want 1", got)
	}
}

func TestInsertNested(t *testing.T) {
	t.Parallel()

	tree := collection.New()
	tree.Insert("1", collection.Root, "API")

	if err := tree.Insert("2", "1", "Users"); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	if !tree.IsBranch("1") {
		t.Error("the parent should be a branch")
	}

	if tree.IsBranch("2") {
		t.Error("a childless node should not be a branch")
	}

	if got := tree.Depth("2"); got != 2 {
		t.Errorf("depth = %d, want 2", got)
	}
}

func TestDuplicateNamesAreAllowedAmongSiblings(t *testing.T) {
	t.Parallel()

	tree := collection.New()

	if err := tree.Insert("1", collection.Root, "Users"); err != nil {
		t.Fatalf("first: %v", err)
	}

	if err := tree.Insert("2", collection.Root, "Users"); err != nil {
		t.Errorf("a duplicate sibling name should be allowed, got %v", err)
	}

	if got := tree.Len(); got != 2 {
		t.Errorf("Len = %d, want both", got)
	}
}

func TestInsertRejectsBadInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		id     string
		parent string
		given  string
		want   error
	}{
		{"empty name", "1", collection.Root, "", collection.ErrEmptyName},
		{"blank name", "1", collection.Root, "  \t ", collection.ErrEmptyName},
		{"control characters", "1", collection.Root, "a\x00b", collection.ErrInvalidName},
		{"unknown parent", "1", "nope", "Users", collection.ErrParentUnknown},
		{"empty identifier", collection.Root, collection.Root, "Users", collection.ErrDuplicateID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tree := collection.New()

			if err := tree.Insert(tt.id, tt.parent, tt.given); !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}

			if got := tree.Len(); got != 0 {
				t.Errorf("tree grew to %d despite the error", got)
			}
		})
	}
}

func TestInsertRejectsARepeatedIdentifier(t *testing.T) {
	t.Parallel()

	tree := collection.New()
	tree.Insert("1", collection.Root, "API")

	if err := tree.Insert("1", collection.Root, "Other"); !errors.Is(err, collection.ErrDuplicateID) {
		t.Errorf("err = %v, want ErrDuplicateID", err)
	}
}

func TestNestingStopsAtMaxDepth(t *testing.T) {
	t.Parallel()

	tree := collection.New()
	parent := collection.Root

	for level := 1; level <= collection.MaxDepth; level++ {
		id := string(rune('a' + level))

		if err := tree.Insert(id, parent, "level"); err != nil {
			t.Fatalf("level %d: %v", level, err)
		}

		parent = id
	}

	if tree.CanAddTo(parent) {
		t.Errorf("CanAddTo at depth %d should be false", collection.MaxDepth)
	}

	if err := tree.Insert("z", parent, "too deep"); !errors.Is(err, collection.ErrTooDeep) {
		t.Errorf("err = %v, want ErrTooDeep", err)
	}
}

func TestNameIsTrimmed(t *testing.T) {
	t.Parallel()

	tree := collection.New()
	tree.Insert("1", collection.Root, "  Padded  ")

	if name, _ := tree.Name("1"); name != "Padded" {
		t.Errorf("name = %q", name)
	}
}

func TestChildrenIsACopy(t *testing.T) {
	t.Parallel()

	tree := collection.New()
	tree.Insert("1", collection.Root, "API")

	roots := tree.Children(collection.Root)
	roots[0] = "tampered"

	if got := tree.Children(collection.Root); got[0] != "1" {
		t.Error("Children exposed the internal slice")
	}
}

func TestUnknownNode(t *testing.T) {
	t.Parallel()

	tree := collection.New()

	if _, ok := tree.Name("ghost"); ok {
		t.Error("Name found a node that does not exist")
	}

	if got := tree.Children("ghost"); got != nil {
		t.Errorf("Children(ghost) = %v, want nil", got)
	}

	if tree.CanAddTo("ghost") {
		t.Error("CanAddTo accepted an unknown parent")
	}
}
