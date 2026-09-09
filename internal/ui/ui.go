package ui

import (
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/GGermanBoldyrev/curlew/internal/collection"
	"github.com/GGermanBoldyrev/curlew/internal/i18n"
	"github.com/GGermanBoldyrev/curlew/internal/paths"
	"github.com/GGermanBoldyrev/curlew/internal/rest"
	"github.com/GGermanBoldyrev/curlew/internal/settings"
)

const (
	dialogWidth = 520
	gutter      = 16
	gap         = 8
)

type CollectionStore interface {
	Load() (*collection.Tree, error)
	Create(parent, name string) (string, error)
	Rename(id, name string) error
	Remove(id string) error
}

type Deps struct {
	Dirs        paths.Dirs
	Log         *slog.Logger
	Settings    *settings.Registry
	T           *i18n.Translator
	Collections *collection.Tree
	Store       CollectionStore
}

type UI struct {
	win      fyne.Window
	prefs    fyne.Preferences
	settings *settings.Registry
	t        *i18n.Translator
	log      *slog.Logger

	collections *collection.Tree
	store       CollectionStore
	tree        *widget.Tree
	method      *widget.Select
	url         *widget.Entry
	send        *widget.Button
	meta        *widget.Label
	split       *container.Split
	editorSplit *container.Split
	request     *rest.Request
	actions     *fyne.Container
	add         *widget.Button
	nest        *widget.Button
	rename      *widget.Button
	remove      *widget.Button
	selected    string
	opened      map[string]bool
}

func New(a fyne.App, d Deps) *UI {
	u := &UI{
		win:         a.NewWindow(""),
		prefs:       a.Preferences(),
		settings:    d.Settings,
		t:           d.T,
		log:         d.Log,
		collections: d.Collections,
		store:       d.Store,
		opened:      make(map[string]bool),
		request:     rest.NewRequest(),
	}

	u.win.SetContent(u.content())
	u.win.Resize(restoreSize(u.prefs))
	u.win.CenterOnScreen()
	u.win.SetMaster()
	u.win.SetOnClosed(u.remember)
	u.win.Canvas().SetOnTypedKey(u.onKey)

	u.log.Debug("window created", slog.String("collections_dir", d.Dirs.Collections()))

	return u
}

func (u *UI) Window() fyne.Window {
	return u.win
}

func (u *UI) Rebuild() {
	u.win.SetContent(u.content())
}

func pad(content fyne.CanvasObject) fyne.CanvasObject {
	return container.New(layout.NewCustomPaddedLayout(gutter, gutter, gutter, gutter), content)
}

func present(d dialog.Dialog) {
	d.Resize(fyne.NewSize(dialogWidth, d.MinSize().Height))
	d.Show()
}

func (u *UI) content() fyne.CanvasObject {
	u.split = container.NewHSplit(u.sidebar(), u.workspace())
	u.split.SetOffset(restoreSidebar(u.prefs))

	return container.NewBorder(u.header(), nil, nil, nil, u.split)
}

func (u *UI) onKey(event *fyne.KeyEvent) {
	if event.Name == fyne.KeyEscape {
		u.deselect()
	}
}

func (u *UI) remember() {
	saveSize(u.prefs, u.win.Canvas().Size())

	if u.split != nil {
		saveSidebar(u.prefs, u.split.Offset)
	}

	if u.editorSplit != nil {
		saveEditor(u.prefs, u.editorSplit.Offset)
	}
}
