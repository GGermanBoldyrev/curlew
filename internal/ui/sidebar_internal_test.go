package ui

import (
	"slices"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/GGermanBoldyrev/curlew/internal/collection"
)

func newSidebar(t *testing.T) (*UI, *collection.Store) {
	t.Helper()

	deps := testDeps(t)

	store, ok := deps.Store.(*collection.Store)
	if !ok {
		t.Fatal("test deps carry an unexpected store")
	}

	return New(test.NewTempApp(t), deps), store
}

func hasTree(object fyne.CanvasObject) bool {
	switch typed := object.(type) {
	case *widget.Tree:
		return true
	case *fyne.Container:
		return slices.ContainsFunc(typed.Objects, hasTree)
	}

	return false
}

func visibleActions(u *UI) int {
	count := 0

	for _, object := range u.actions.Objects {
		if object.Visible() {
			count++
		}
	}

	return count
}

func rootIDs(t *testing.T, u *UI) []string {
	t.Helper()

	return u.collections.Children(collection.Root)
}

func TestSidebarShowsTheHintUntilTheFirstCollection(t *testing.T) {
	u, _ := newSidebar(t)

	if hasTree(u.body()) {
		t.Fatal("an empty sidebar should show the hint, not the tree")
	}

	u.addCollection(collection.Root, "Payments API")

	if !hasTree(u.body()) {
		t.Error("the tree should replace the hint once a collection exists")
	}
}

func TestAddedCollectionReachesTheDatabase(t *testing.T) {
	u, store := newSidebar(t)

	u.addCollection(collection.Root, "Payments API")
	api := rootIDs(t, u)[0]
	u.addCollection(api, "Users")

	tree, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := tree.Len(); got != 2 {
		t.Fatalf("stored = %d, want 2", got)
	}

	children := tree.Children(api)
	if len(children) != 1 {
		t.Fatalf("children = %v", children)
	}

	if name, _ := tree.Name(children[0]); name != "Users" {
		t.Errorf("child = %q", name)
	}
}

func TestDuplicateNamesAreAccepted(t *testing.T) {
	u, _ := newSidebar(t)

	u.addCollection(collection.Root, "Users")
	u.addCollection(collection.Root, "Users")

	if got := u.collections.Len(); got != 2 {
		t.Errorf("collections = %d, want duplicates allowed", got)
	}
}

func TestBlankNameIsRejectedAndExplained(t *testing.T) {
	u, _ := newSidebar(t)

	u.addCollection(collection.Root, "   ")

	if got := u.collections.Len(); got != 0 {
		t.Errorf("collections = %d, want the blank name rejected", got)
	}

	if u.win.Canvas().Overlays().Top() == nil {
		t.Error("the user was not told why nothing happened")
	}
}

func TestClickingACollectionTogglesIt(t *testing.T) {
	u, _ := newSidebar(t)

	u.addCollection(collection.Root, "API")
	api := rootIDs(t, u)[0]
	u.addCollection(api, "Users")

	u.tree.CloseBranch(api)

	u.pick(api)

	if !u.tree.IsBranchOpen(api) {
		t.Fatal("the first click should open the branch")
	}

	if u.selected != api {
		t.Fatalf("selected = %q, want the clicked collection", u.selected)
	}

	u.pick(api)

	if u.tree.IsBranchOpen(api) {
		t.Error("the second click should close the branch")
	}
}

func TestClickingALeafOnlySelectsIt(t *testing.T) {
	u, _ := newSidebar(t)

	u.addCollection(collection.Root, "API")
	api := rootIDs(t, u)[0]

	u.pick(api)
	u.pick(api)

	if u.selected != api {
		t.Errorf("selected = %q, want the leaf to stay selected", u.selected)
	}
}

func TestActionsAppearOnlyForWhatIsPossible(t *testing.T) {
	u, _ := newSidebar(t)

	if got := visibleActions(u); got != 1 {
		t.Errorf("with nothing selected only the root add applies, got %d", got)
	}

	u.addCollection(collection.Root, "API")
	u.pick(rootIDs(t, u)[0])

	if got := visibleActions(u); got != 4 {
		t.Errorf("with a selection all four actions apply, got %d", got)
	}
}

func TestDeepestCollectionLosesTheAddAction(t *testing.T) {
	u, _ := newSidebar(t)

	parent := collection.Root
	for range collection.MaxDepth {
		u.addCollection(parent, "level")
		parent = u.collections.Children(parent)[0]
	}

	u.pick(parent)

	if got := visibleActions(u); got != 3 {
		t.Errorf("at the deepest level nesting is gone, got %d actions", got)
	}
}

func TestEscapeClearsTheSelection(t *testing.T) {
	u, _ := newSidebar(t)

	u.addCollection(collection.Root, "API")
	u.pick(rootIDs(t, u)[0])

	if u.selected == collection.Root {
		t.Fatal("nothing was selected to begin with")
	}

	u.onKey(&fyne.KeyEvent{Name: fyne.KeyEscape})

	if u.selected != collection.Root {
		t.Errorf("selected = %q, want it cleared so a root collection can be added", u.selected)
	}

	if got := visibleActions(u); got != 1 {
		t.Errorf("after clearing only the root add should remain, got %d", got)
	}
}

func TestRenameIsPersisted(t *testing.T) {
	u, store := newSidebar(t)

	u.addCollection(collection.Root, "API")
	api := rootIDs(t, u)[0]
	u.addCollection(api, "Users")

	u.renameCollection(api, "Payments API")

	tree, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if name, _ := tree.Name(api); name != "Payments API" {
		t.Errorf("name = %q", name)
	}

	if got := len(tree.Children(api)); got != 1 {
		t.Errorf("the rename lost the children: %d", got)
	}
}

func TestRemovalIsPersistedAndCascades(t *testing.T) {
	u, store := newSidebar(t)

	u.addCollection(collection.Root, "API")
	api := rootIDs(t, u)[0]
	u.addCollection(api, "Users")
	u.addCollection(collection.Root, "Internal")

	u.pick(api)
	u.removeCollection(api)

	tree, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := tree.Len(); got != 1 {
		t.Errorf("stored = %d, want only the sibling", got)
	}

	if u.selected != collection.Root {
		t.Errorf("selection = %q, want it cleared", u.selected)
	}
}

func TestRootAddIgnoresTheSelection(t *testing.T) {
	u, _ := newSidebar(t)

	u.addCollection(collection.Root, "API")
	api := rootIDs(t, u)[0]
	u.pick(api)

	u.promptForRoot()

	if got := len(u.collections.Children(api)); got != 0 {
		t.Errorf("the root action must not nest, API gained %d children", got)
	}
}
