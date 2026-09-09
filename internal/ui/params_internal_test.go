package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/GGermanBoldyrev/curlew/internal/rest"
)

func collect[T fyne.CanvasObject](object fyne.CanvasObject, found *[]T) {
	if typed, ok := object.(T); ok {
		*found = append(*found, typed)
	}

	switch typed := object.(type) {
	case *fyne.Container:
		for _, child := range typed.Objects {
			collect(child, found)
		}
	case *container.Scroll:
		collect(typed.Content, found)
	case *container.Split:
		collect(typed.Leading, found)
		collect(typed.Trailing, found)
	case *container.AppTabs:
		for _, item := range typed.Items {
			collect(item.Content, found)
		}
	}
}

func find[T fyne.CanvasObject](object fyne.CanvasObject) []T {
	var found []T
	collect(object, &found)

	return found
}

func TestParamRowsWriteIntoTheRequest(t *testing.T) {
	u := New(test.NewTempApp(t), testDeps(t))
	rows := []rest.Param{{Key: "page", Value: "1"}}

	table := u.paramTable(&rows, "kv.empty.params")

	entries := find[*widget.Entry](table)
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want key and value", len(entries))
	}

	entries[0].SetText("limit")
	entries[1].SetText("50")

	if rows[0].Key != "limit" || rows[0].Value != "50" {
		t.Errorf("row = %+v, want the edits applied", rows[0])
	}
}

func TestDisablingARowFlipsTheModel(t *testing.T) {
	u := New(test.NewTempApp(t), testDeps(t))
	rows := []rest.Param{{Key: "page", Value: "1"}}

	table := u.paramTable(&rows, "kv.empty.params")

	checks := find[*widget.Check](table)
	if len(checks) != 1 {
		t.Fatalf("checks = %d, want one", len(checks))
	}

	if !checks[0].Checked {
		t.Fatal("an enabled row should start checked")
	}

	checks[0].SetChecked(false)

	if !rows[0].Disabled {
		t.Error("unchecking the row did not disable it")
	}
}

func TestAddingAndRemovingRows(t *testing.T) {
	u := New(test.NewTempApp(t), testDeps(t))

	var rows []rest.Param

	table := u.paramTable(&rows, "kv.empty.params")

	buttons := find[*widget.Button](table)
	if len(buttons) != 1 {
		t.Fatalf("buttons = %d, want only the add button on an empty table", len(buttons))
	}

	test.Tap(buttons[0])

	if len(rows) != 1 {
		t.Fatalf("rows = %d, want one after adding", len(rows))
	}

	buttons = find[*widget.Button](table)
	if len(buttons) != 2 {
		t.Fatalf("buttons = %d, want add and delete", len(buttons))
	}

	for _, button := range buttons {
		if button.Text == "" {
			test.Tap(button)

			break
		}
	}

	if len(rows) != 0 {
		t.Errorf("rows = %d, want the row removed", len(rows))
	}
}

func TestBodyKindSwitchesTheModel(t *testing.T) {
	u := New(test.NewTempApp(t), testDeps(t))

	body := u.bodyTab()

	selects := find[*widget.Select](body)
	if len(selects) != 1 {
		t.Fatalf("selects = %d, want the body kind", len(selects))
	}

	selects[0].SetSelected(u.t.T("body.json"))

	if u.request.Body.Kind != rest.BodyRaw {
		t.Errorf("kind = %v, want raw", u.request.Body.Kind)
	}

	if u.request.Body.ContentType != jsonContentType {
		t.Errorf("content type = %q, want %q", u.request.Body.ContentType, jsonContentType)
	}

	selects[0].SetSelected(u.t.T("body.none"))

	if u.request.Body.Kind != rest.BodyNone {
		t.Errorf("kind = %v, want none", u.request.Body.Kind)
	}
}

func TestRequestSurvivesARebuild(t *testing.T) {
	u := New(test.NewTempApp(t), testDeps(t))

	u.request.URL = "https://example.com"
	u.request.Method = rest.MethodPost
	u.request.Query = []rest.Param{{Key: "page", Value: "2"}}

	u.Rebuild()

	if u.url.Text != "https://example.com" {
		t.Errorf("url = %q, want it restored", u.url.Text)
	}

	if u.method.Selected != rest.MethodPost {
		t.Errorf("method = %q, want it restored", u.method.Selected)
	}

	if len(u.request.Query) != 1 {
		t.Errorf("query = %v, want it kept", u.request.Query)
	}
}
