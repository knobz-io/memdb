package memdb

type CondFunc[V any] func(V) bool

func (fn CondFunc[V]) Matches(v V) bool {
	return fn(v)
}

type Cond[V any] interface {
	Matches(v V) bool
}
