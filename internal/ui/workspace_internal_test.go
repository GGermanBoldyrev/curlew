package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"github.com/GGermanBoldyrev/curlew/internal/rest"
)

func TestWorkspaceCarriesTheRequestBar(t *testing.T) {
	u := New(test.NewTempApp(t), testDeps(t))

	if u.method == nil || u.url == nil || u.send == nil {
		t.Fatal("the request bar is incomplete")
	}

	if got := u.method.Selected; got != rest.MethodGet {
		t.Errorf("default method = %q, want %q", got, rest.MethodGet)
	}

	if got := len(u.method.Options); got != len(rest.Methods()) {
		t.Errorf("method options = %d, want %d", got, len(rest.Methods()))
	}

	if !u.send.Disabled() {
		t.Error("Send is enabled although nothing can be sent yet")
	}
}

func TestWorkspaceSurvivesALanguageRebuild(t *testing.T) {
	u := New(test.NewTempApp(t), testDeps(t))

	before := u.send.Text

	u.t.SetLang("ru")
	u.Rebuild()

	if u.send.Text == before {
		t.Errorf("the Send button was not relabelled, still %q", u.send.Text)
	}

	if u.url.PlaceHolder == "" {
		t.Error("the url placeholder went missing after the rebuild")
	}
}
