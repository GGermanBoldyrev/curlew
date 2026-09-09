package rest

type Inherited[T any] struct {
	value T
	set   bool
}

func Value[T any](v T) Inherited[T] { return Inherited[T]{value: v, set: true} }

func Inherit[T any]() Inherited[T] { return Inherited[T]{} }

func (i Inherited[T]) Get() (T, bool) { return i.value, i.set }

func (i Inherited[T]) IsSet() bool { return i.set }

func (i Inherited[T]) Or(parent Inherited[T]) Inherited[T] {
	if i.set {
		return i
	}
	return parent
}

func (i Inherited[T]) OrValue(def T) T {
	if i.set {
		return i.value
	}
	return def
}
