package rest_test

import (
	"testing"
	"time"

	"github.com/GGermanBoldyrev/curlew/internal/rest"
)

func TestInheritedOr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		child   rest.Inherited[bool]
		parent  rest.Inherited[bool]
		want    bool
		wantSet bool
	}{
		{"child set wins", rest.Value(false), rest.Value(true), false, true},
		{"child unset falls through", rest.Inherit[bool](), rest.Value(true), true, true},
		{"both unset stays unset", rest.Inherit[bool](), rest.Inherit[bool](), false, false},
		{"explicit false is not absent", rest.Value(false), rest.Inherit[bool](), false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, set := tt.child.Or(tt.parent).Get()
			if got != tt.want || set != tt.wantSet {
				t.Errorf("Or() = (%v, %v), want (%v, %v)", got, set, tt.want, tt.wantSet)
			}
		})
	}
}

func TestInheritedOrValue(t *testing.T) {
	t.Parallel()

	if got := rest.Inherit[time.Duration]().OrValue(30 * time.Second); got != 30*time.Second {
		t.Errorf("unset OrValue = %v, want 30s", got)
	}
	if got := rest.Value(5 * time.Second).OrValue(30 * time.Second); got != 5*time.Second {
		t.Errorf("set OrValue = %v, want 5s", got)
	}
}

func TestOptionsOr(t *testing.T) {
	t.Parallel()

	collection := rest.Options{
		Timeout:         rest.Value(30 * time.Second),
		FollowRedirects: rest.Value(true),
	}
	folder := rest.Options{FollowRedirects: rest.Value(false)}
	request := rest.Options{Timeout: rest.Value(5 * time.Second)}

	got := request.Or(folder).Or(collection)

	if v := got.Timeout.OrValue(0); v != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s from the request", v)
	}
	if v := got.FollowRedirects.OrValue(true); v {
		t.Error("FollowRedirects = true, want false from the folder")
	}
	if got.Proxy.IsSet() {
		t.Error("Proxy should stay unset when nothing in the chain sets it")
	}
}
