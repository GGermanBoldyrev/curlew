package settings_test

import (
	"testing"

	"github.com/GGermanBoldyrev/curlew/internal/settings"
)

type fakeStore map[string]string

func (f fakeStore) StringWithFallback(key, fallback string) string {
	if v, ok := f[key]; ok {
		return v
	}

	return fallback
}

func (f fakeStore) SetString(key, value string) {
	f[key] = value
}

func theme() settings.Option {
	return settings.Option{Key: "theme", Default: "dark", Values: []settings.Value{"light", "dark"}}
}

func TestOptionKeys(t *testing.T) {
	t.Parallel()

	o := theme()

	if got := o.PrefKey(); got != "settings.theme" {
		t.Errorf("PrefKey = %q", got)
	}

	if got := o.NameKey(); got != "settings.theme.name" {
		t.Errorf("NameKey = %q", got)
	}

	if got := o.ValueKey("light"); got != "settings.theme.light" {
		t.Errorf("ValueKey = %q", got)
	}
}

func TestGetReturnsDefaultForUnknownStoredValue(t *testing.T) {
	t.Parallel()

	o := theme()
	store := fakeStore{o.PrefKey(): "chartreuse"}
	r := settings.New(store, o)

	if got := r.Get(o); got != o.Default {
		t.Errorf("Get = %q, want the default %q", got, o.Default)
	}
}

func TestSetPersistsAndNotifies(t *testing.T) {
	t.Parallel()

	o := theme()
	store := fakeStore{}
	r := settings.New(store, o)

	var seen []settings.Value
	r.OnChange(func(_ settings.Option, v settings.Value) { seen = append(seen, v) })

	r.Set(o, "light")

	if got := r.Get(o); got != "light" {
		t.Errorf("Get = %q", got)
	}

	if store[o.PrefKey()] != "light" {
		t.Errorf("store = %q", store[o.PrefKey()])
	}

	if len(seen) != 1 || seen[0] != "light" {
		t.Errorf("notifications = %v, want one 'light'", seen)
	}
}

func TestSetIgnoresUnchangedAndDisallowedValues(t *testing.T) {
	t.Parallel()

	o := theme()
	r := settings.New(fakeStore{}, o)

	var calls int
	r.OnChange(func(_ settings.Option, _ settings.Value) { calls++ })

	r.Set(o, o.Default)
	r.Set(o, "chartreuse")

	if calls != 0 {
		t.Errorf("notifications = %d, want 0", calls)
	}
}

func TestOptionsIsACopy(t *testing.T) {
	t.Parallel()

	o := theme()
	r := settings.New(fakeStore{}, o)

	opts := r.Options()
	opts[0].Key = "mutated"

	if r.Options()[0].Key != "theme" {
		t.Error("Options() exposed the internal slice")
	}
}
