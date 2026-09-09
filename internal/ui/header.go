package ui

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/GGermanBoldyrev/curlew/internal/settings"
)

const appName = "Curlew"

func (u *UI) header() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(appName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	controls := container.NewHBox()
	for _, option := range u.settings.Options() {
		controls.Add(u.optionSelect(option))
	}

	bar := container.NewStack(
		container.NewCenter(title),
		container.NewBorder(nil, nil, nil, controls, layout.NewSpacer()),
	)

	return container.NewVBox(bar, widget.NewSeparator())
}

func (u *UI) optionSelect(option settings.Option) fyne.CanvasObject {
	labels := make([]string, 0, len(option.Values))
	for _, value := range option.Values {
		labels = append(labels, u.t.T(option.ValueKey(value)))
	}

	sel := widget.NewSelect(labels, nil)
	sel.Selected = u.t.T(option.ValueKey(u.settings.Get(option)))
	sel.OnChanged = func(label string) {
		if i := slices.Index(labels, label); i >= 0 {
			u.settings.Set(option, option.Values[i])
		}
	}

	return sel
}
