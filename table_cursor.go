package memdb

type TableCursor[V any] struct {
	table Table[V]
	ids   *treeTxn[struct{}]
	idx   *treeTxn[V]
	order *treeTxn[*tree[struct{}]]
	dir   OrderDirection
}

func (t *TableCursor[V]) Seek(id Key) (V, bool) {
	return *new(V), false
}

func (t *TableCursor[V]) First() (V, bool) {
	return *new(V), false
}

func (t *TableCursor[V]) Last() (V, bool) {
	return *new(V), false
}

func (t *TableCursor[V]) Next() (V, bool) {
	return *new(V), false
}

func (t *TableCursor[V]) Prev() (V, bool) {
	return *new(V), false
}
