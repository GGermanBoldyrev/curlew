package settings

import (
	"slices"
	"sync"
)

const (
	KeyLanguage = "language"
	KeyTheme    = "theme"
)

type Value string

type Option struct {
	Key     string
	Default Value
	Values  []Value
}

func (o Option) PrefKey() string {
	return "settings." + o.Key
}

func (o Option) NameKey() string {
	return "settings." + o.Key + ".name"
}

func (o Option) ValueKey(v Value) string {
	return "settings." + o.Key + "." + string(v)
}

func (o Option) Allows(v Value) bool {
	return slices.Contains(o.Values, v)
}

type Store interface {
	StringWithFallback(key, fallback string) string
	SetString(key, value string)
}

type Registry struct {
	mu    sync.RWMutex
	store Store
	opts  []Option
	subs  []func(Option, Value)
}

func New(store Store, opts ...Option) *Registry {
	return &Registry{store: store, opts: slices.Clone(opts)}
}

func (r *Registry) Options() []Option {
	return slices.Clone(r.opts)
}

func (r *Registry) Get(o Option) Value {
	v := Value(r.store.StringWithFallback(o.PrefKey(), string(o.Default)))
	if !o.Allows(v) {
		return o.Default
	}

	return v
}

func (r *Registry) Set(o Option, v Value) {
	if !o.Allows(v) || r.Get(o) == v {
		return
	}

	r.store.SetString(o.PrefKey(), string(v))

	r.mu.RLock()
	subs := slices.Clone(r.subs)
	r.mu.RUnlock()

	for _, notify := range subs {
		notify(o, v)
	}
}

func (r *Registry) OnChange(notify func(Option, Value)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.subs = append(r.subs, notify)
}
