package collection

import (
	"errors"
	"slices"
	"strings"
	"unicode"
)

const (
	Root = ""

	MaxDepth = 5
)

var (
	ErrEmptyName     = errors.New("collection name must not be empty")
	ErrInvalidName   = errors.New("collection name contains characters that are not allowed")
	ErrTooDeep       = errors.New("maximum nesting depth reached")
	ErrParentUnknown = errors.New("parent collection not found")
	ErrDuplicateID   = errors.New("collection identifier is already taken")
	ErrNotFound      = errors.New("collection not found")
)

type Tree struct {
	nodes map[string]*node
	roots []string
}

type node struct {
	name     string
	parent   string
	children []string
}

func New() *Tree {
	return &Tree{nodes: make(map[string]*node)}
}

func (t *Tree) Insert(id, parent, name string) error {
	if id == Root {
		return ErrDuplicateID
	}

	if _, taken := t.nodes[id]; taken {
		return ErrDuplicateID
	}

	name, err := clean(name)
	if err != nil {
		return err
	}

	if parent != Root {
		if _, ok := t.nodes[parent]; !ok {
			return ErrParentUnknown
		}
	}

	if !t.CanAddTo(parent) {
		return ErrTooDeep
	}

	t.nodes[id] = &node{name: name, parent: parent}

	if parent == Root {
		t.roots = append(t.roots, id)

		return nil
	}

	t.nodes[parent].children = append(t.nodes[parent].children, id)

	return nil
}

func (t *Tree) Children(id string) []string {
	if id == Root {
		return slices.Clone(t.roots)
	}

	found, ok := t.nodes[id]
	if !ok {
		return nil
	}

	return slices.Clone(found.children)
}

func (t *Tree) IsBranch(id string) bool {
	return len(t.Children(id)) > 0
}

func (t *Tree) Name(id string) (string, bool) {
	found, ok := t.nodes[id]
	if !ok {
		return "", false
	}

	return found.name, true
}

func (t *Tree) Depth(id string) int {
	depth := 0

	for id != Root {
		found, ok := t.nodes[id]
		if !ok {
			return 0
		}

		depth++
		id = found.parent
	}

	return depth
}

func (t *Tree) CanAddTo(parent string) bool {
	if parent != Root {
		if _, ok := t.nodes[parent]; !ok {
			return false
		}
	}

	return t.Depth(parent) < MaxDepth
}

func (t *Tree) Len() int {
	return len(t.nodes)
}

func clean(name string) (string, error) {
	name = strings.TrimSpace(name)

	return name, ValidateName(name)
}

func ValidateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}

	if strings.ContainsFunc(name, unicode.IsControl) {
		return ErrInvalidName
	}

	return nil
}
