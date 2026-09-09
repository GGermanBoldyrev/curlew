package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/GGermanBoldyrev/curlew/internal/rest"
)

func (u *UI) workspace() fyne.CanvasObject {
	editor := container.NewBorder(u.requestBar(), nil, nil, nil, u.requestTabs())

	u.editorSplit = container.NewVSplit(editor, u.responsePane())
	u.editorSplit.SetOffset(restoreEditor(u.prefs))

	return u.editorSplit
}

func (u *UI) requestTabs() fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem(u.t.T("tab.params"), pad(u.paramTable(&u.request.Query, "kv.empty.params"))),
		container.NewTabItem(u.t.T("tab.headers"), pad(u.paramTable(&u.request.Headers, "kv.empty.headers"))),
		container.NewTabItem(u.t.T("tab.body"), pad(u.bodyTab())),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

func (u *UI) requestBar() fyne.CanvasObject {
	u.method = widget.NewSelect(rest.Methods(), func(chosen string) { u.request.Method = chosen })
	u.method.Selected = u.request.Method

	u.url = widget.NewEntry()
	u.url.SetPlaceHolder(u.t.T("request.url.placeholder"))
	u.url.SetText(u.request.URL)
	u.url.OnChanged = func(text string) { u.request.URL = text }

	u.send = widget.NewButtonWithIcon(u.t.T("request.send"), theme.MailSendIcon(), nil)
	u.send.Importance = widget.HighImportance
	u.send.Disable()

	return container.NewVBox(
		pad(container.NewBorder(nil, nil, u.method, u.send, u.url)),
		widget.NewSeparator(),
	)
}

func (u *UI) responsePane() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(u.t.T("response.title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	u.meta = widget.NewLabel("")
	u.meta.TextStyle = fyne.TextStyle{Monospace: true}

	format := widget.NewLabel(u.t.T("response.format.json"))
	format.Importance = widget.LowImportance

	head := container.NewBorder(nil, nil, title,
		container.New(layout.NewCustomPaddedHBoxLayout(gap), u.meta, format))

	return container.NewBorder(
		container.NewVBox(pad(head), widget.NewSeparator()),
		nil, nil, nil,
		u.responseBody(),
	)
}

func (u *UI) responseBody() fyne.CanvasObject {
	empty := widget.NewLabel(u.t.T("response.empty"))
	empty.Alignment = fyne.TextAlignCenter
	empty.Wrapping = fyne.TextWrapWord

	return container.NewCenter(empty)
}
