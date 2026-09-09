package ui

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/GGermanBoldyrev/curlew/internal/rest"
)

func (u *UI) paramTable(rows *[]rest.Param, emptyKey string) fyne.CanvasObject {
	list := container.New(layout.NewCustomPaddedVBoxLayout(gap))
	empty := widget.NewLabel(u.t.T(emptyKey))

	var refill func()

	refill = func() {
		list.Objects = nil

		if len(*rows) == 0 {
			list.Add(empty)
		}

		for i := range *rows {
			list.Add(u.paramRow(rows, i, refill))
		}

		list.Refresh()
	}

	refill()

	add := widget.NewButtonWithIcon(u.t.T("kv.add"), theme.ContentAddIcon(), func() {
		*rows = append(*rows, rest.Param{})
		refill()
	})
	add.Importance = widget.LowImportance

	return container.NewBorder(
		nil,
		container.NewHBox(add),
		nil, nil,
		container.NewVScroll(list),
	)
}

func (u *UI) paramRow(rows *[]rest.Param, index int, refill func()) fyne.CanvasObject {
	row := &(*rows)[index]

	enabled := widget.NewCheck("", func(checked bool) { row.Disabled = !checked })
	enabled.Checked = !row.Disabled

	key := widget.NewEntry()
	key.SetPlaceHolder(u.t.T("kv.key"))
	key.SetText(row.Key)
	key.OnChanged = func(value string) { row.Key = value }

	value := widget.NewEntry()
	value.SetPlaceHolder(u.t.T("kv.value"))
	value.SetText(row.Value)
	value.OnChanged = func(text string) { row.Value = text }

	drop := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		*rows = slices.Delete(*rows, index, index+1)
		refill()
	})
	drop.Importance = widget.LowImportance

	return container.NewBorder(nil, nil, enabled, drop, container.NewGridWithColumns(2, key, value))
}
