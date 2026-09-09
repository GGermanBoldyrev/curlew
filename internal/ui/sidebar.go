package ui

import (
	"errors"
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/GGermanBoldyrev/curlew/internal/collection"
)

func (u *UI) sidebar() fyne.CanvasObject {
	u.tree = widget.NewTree(
		u.collections.Children,
		u.collections.IsBranch,
		func(bool) fyne.CanvasObject { return newNode(u.pick) },
		u.showNode,
	)

	u.tree.OnSelected = func(id widget.TreeNodeID) {
		u.selected = id
		u.refreshActions()
	}
	u.tree.OnUnselected = func(widget.TreeNodeID) {
		u.selected = collection.Root
		u.refreshActions()
	}
	u.tree.OnBranchOpened = func(id widget.TreeNodeID) { u.opened[id] = true }
	u.tree.OnBranchClosed = func(id widget.TreeNodeID) { delete(u.opened, id) }

	for id := range u.opened {
		u.tree.OpenBranch(id)
	}

	if _, ok := u.collections.Name(u.selected); ok {
		u.tree.Select(u.selected)
	} else {
		u.selected = collection.Root
	}

	u.add = u.action(theme.ContentAddIcon(), u.promptForRoot)
	u.nest = u.action(theme.FolderNewIcon(), u.promptForChild)
	u.rename = u.action(theme.DocumentCreateIcon(), u.promptForRename)
	u.remove = u.action(theme.DeleteIcon(), u.confirmRemoval)
	u.actions = container.NewHBox(u.add, u.nest, u.rename, u.remove)

	u.refreshActions()

	title := widget.NewLabelWithStyle(u.t.T("sidebar.title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Truncation = fyne.TextTruncateEllipsis

	head := container.NewBorder(nil, nil, nil, u.actions, title)

	return container.NewBorder(head, nil, nil, nil, u.body())
}

func (u *UI) body() fyne.CanvasObject {
	if u.collections.Len() == 0 {
		empty := widget.NewLabel(u.t.T("sidebar.empty"))
		empty.Alignment = fyne.TextAlignCenter
		empty.Wrapping = fyne.TextWrapWord

		return container.NewCenter(empty)
	}

	return u.tree
}

func (u *UI) deselect() {
	if u.tree != nil {
		u.tree.UnselectAll()
	}

	u.selected = collection.Root
	u.refreshActions()
}

func (u *UI) refreshActions() {
	if u.actions == nil {
		return
	}

	chosen := u.selected != collection.Root

	show(u.nest, chosen && u.collections.CanAddTo(u.selected))
	show(u.rename, chosen)
	show(u.remove, chosen)

	u.actions.Refresh()
}

func show(object fyne.CanvasObject, visible bool) {
	if visible {
		object.Show()

		return
	}

	object.Hide()
}

func (u *UI) action(icon fyne.Resource, tapped func()) *widget.Button {
	button := widget.NewButtonWithIcon("", icon, tapped)
	button.Importance = widget.LowImportance

	return button
}

func (u *UI) showNode(id widget.TreeNodeID, _ bool, object fyne.CanvasObject) {
	node, ok := object.(*node)
	if !ok {
		return
	}

	name, _ := u.collections.Name(id)
	node.show(id, name)
}

func (u *UI) pick(id string) {
	switch {
	case !u.collections.IsBranch(id):
	case u.selected == id && u.tree.IsBranchOpen(id):
		u.tree.CloseBranch(id)
	default:
		u.tree.OpenBranch(id)
	}

	u.tree.Select(id)
	u.selected = id
	u.refreshActions()
}

func (u *UI) promptForRoot() {
	u.askForName(u.t.T("collection.new.root"), u.t.T("dialog.add"), "", func(name string) {
		u.addCollection(collection.Root, name)
	})
}

func (u *UI) promptForChild() {
	parent := u.selected

	name, ok := u.collections.Name(parent)
	if !ok {
		return
	}

	u.askForName(fmt.Sprintf(u.t.T("collection.new.child"), name),
		u.t.T("dialog.add"), "", func(chosen string) {
			u.addCollection(parent, chosen)
		})
}

func (u *UI) promptForRename() {
	id := u.selected

	current, ok := u.collections.Name(id)
	if !ok {
		return
	}

	u.askForName(fmt.Sprintf(u.t.T("collection.rename.title"), current),
		u.t.T("dialog.rename"), current, func(name string) {
			u.renameCollection(id, name)
		})
}

func (u *UI) confirmRemoval() {
	id := u.selected

	name, ok := u.collections.Name(id)
	if !ok {
		return
	}

	message := widget.NewLabel(fmt.Sprintf(u.t.T("collection.delete.message"), name))
	message.Wrapping = fyne.TextWrapWord

	confirm := dialog.NewCustomConfirm(
		u.t.T("collection.delete.title"),
		u.t.T("dialog.delete"),
		u.t.T("dialog.cancel"),
		pad(message),
		func(confirmed bool) {
			if confirmed {
				u.removeCollection(id)
			}
		},
		u.win,
	)

	confirm.SetConfirmImportance(widget.DangerImportance)

	present(confirm)
}

func (u *UI) askForName(title, confirm, initial string, apply func(string)) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(u.t.T("collection.name.placeholder"))
	entry.SetText(initial)

	caption := widget.NewLabelWithStyle(u.t.T("collection.name"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	body := container.New(layout.NewCustomPaddedVBoxLayout(gap), caption, entry)

	form := dialog.NewCustomConfirm(
		title,
		confirm,
		u.t.T("dialog.cancel"),
		pad(body),
		func(confirmed bool) {
			if confirmed {
				apply(entry.Text)
			}
		},
		u.win,
	)

	present(form)

	u.win.Canvas().Focus(entry)
}

func (u *UI) addCollection(parent, name string) {
	if _, err := u.store.Create(parent, name); err != nil {
		u.report(err)

		return
	}

	u.opened[parent] = true
	u.reload()
}

func (u *UI) renameCollection(id, name string) {
	if err := u.store.Rename(id, name); err != nil {
		u.report(err)

		return
	}

	u.reload()
}

func (u *UI) removeCollection(id string) {
	if err := u.store.Remove(id); err != nil {
		u.report(err)

		return
	}

	delete(u.opened, id)
	u.selected = collection.Root
	u.reload()
}

func (u *UI) reload() {
	tree, err := u.store.Load()
	if err != nil {
		u.report(err)

		return
	}

	u.collections = tree
	u.Rebuild()
}

func (u *UI) report(err error) {
	message := widget.NewLabel(u.explain(err))
	message.Wrapping = fyne.TextWrapWord

	present(dialog.NewCustom(u.t.T("collection.error.title"), u.t.T("dialog.close"), pad(message), u.win))
}

func (u *UI) explain(err error) string {
	switch {
	case errors.Is(err, collection.ErrEmptyName):
		return u.t.T("collection.error.empty")
	case errors.Is(err, collection.ErrInvalidName):
		return u.t.T("collection.error.invalid")
	case errors.Is(err, collection.ErrTooDeep):
		return fmt.Sprintf(u.t.T("collection.error.deep"), collection.MaxDepth)
	}

	u.log.Error("collection storage", slog.Any("error", err))

	return fmt.Sprintf(u.t.T("collection.error.save"), err)
}

type node struct {
	widget.BaseWidget

	text  *widget.Label
	id    string
	onTap func(string)
}

func newNode(onTap func(string)) *node {
	n := &node{text: widget.NewLabel(""), onTap: onTap}
	n.ExtendBaseWidget(n)

	return n
}

func (n *node) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(n.text)
}

func (n *node) Tapped(*fyne.PointEvent) {
	if n.onTap != nil {
		n.onTap(n.id)
	}
}

func (n *node) show(id, name string) {
	n.id = id
	n.text.SetText(name)
}
