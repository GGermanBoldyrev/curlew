package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/GGermanBoldyrev/curlew/internal/rest"
)

const jsonContentType = "application/json"

func (u *UI) bodyTab() fyne.CanvasObject {
	editor := widget.NewMultiLineEntry()
	editor.SetPlaceHolder(u.t.T("body.placeholder"))
	editor.TextStyle = fyne.TextStyle{Monospace: true}
	editor.Wrapping = fyne.TextWrapOff
	editor.SetText(u.request.Body.Text)
	editor.OnChanged = func(text string) { u.request.Body.Text = text }

	kinds := []string{u.t.T("body.none"), u.t.T("body.json")}

	kind := widget.NewSelect(kinds, nil)
	kind.Selected = kinds[bodyIndex(u.request.Body.Kind)]
	kind.OnChanged = func(chosen string) {
		if chosen == kinds[0] {
			u.request.Body.Kind = rest.BodyNone
			u.request.Body.ContentType = ""
			editor.Hide()

			return
		}

		u.request.Body.Kind = rest.BodyRaw
		u.request.Body.ContentType = jsonContentType
		editor.Show()
	}

	if u.request.Body.Kind == rest.BodyNone {
		editor.Hide()
	}

	head := container.New(layout.NewCustomPaddedHBoxLayout(gap),
		widget.NewLabel(u.t.T("body.kind")), kind)

	return container.NewBorder(head, nil, nil, nil, editor)
}

func bodyIndex(kind rest.BodyKind) int {
	if kind == rest.BodyNone {
		return 0
	}

	return 1
}
