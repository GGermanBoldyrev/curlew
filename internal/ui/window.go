package ui

import "fyne.io/fyne/v2"

const (
	prefWindowWidth  = "window.width"
	prefWindowHeight = "window.height"

	defaultWindowWidth  = 1180
	defaultWindowHeight = 760

	minWindowWidth  = 720
	minWindowHeight = 480
	maxWindowSide   = 8192

	prefEditorOffset    = "editor.offset"
	defaultEditorOffset = 0.45
	minEditorOffset     = 0.15
	maxEditorOffset     = 0.85

	prefSidebarOffset    = "sidebar.offset"
	defaultSidebarOffset = 0.22
	minSidebarOffset     = 0.08
	maxSidebarOffset     = 0.6
)

func restoreSize(p fyne.Preferences) fyne.Size {
	return fyne.NewSize(
		sane(p.FloatWithFallback(prefWindowWidth, defaultWindowWidth), minWindowWidth, defaultWindowWidth),
		sane(p.FloatWithFallback(prefWindowHeight, defaultWindowHeight), minWindowHeight, defaultWindowHeight),
	)
}

func saveSize(p fyne.Preferences, size fyne.Size) {
	p.SetFloat(prefWindowWidth, float64(size.Width))
	p.SetFloat(prefWindowHeight, float64(size.Height))
}

func sane(value, lower, fallback float64) float32 {
	if value < lower || value > maxWindowSide {
		return float32(fallback)
	}

	return float32(value)
}

func restoreSidebar(p fyne.Preferences) float64 {
	offset := p.FloatWithFallback(prefSidebarOffset, defaultSidebarOffset)
	if offset < minSidebarOffset || offset > maxSidebarOffset {
		return defaultSidebarOffset
	}

	return offset
}

func saveSidebar(p fyne.Preferences, offset float64) {
	p.SetFloat(prefSidebarOffset, offset)
}

func restoreEditor(p fyne.Preferences) float64 {
	offset := p.FloatWithFallback(prefEditorOffset, defaultEditorOffset)
	if offset < minEditorOffset || offset > maxEditorOffset {
		return defaultEditorOffset
	}

	return offset
}

func saveEditor(p fyne.Preferences, offset float64) {
	p.SetFloat(prefEditorOffset, offset)
}
