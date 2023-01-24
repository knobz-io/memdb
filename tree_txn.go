package memdb

type treeTxn[V any] struct {
	root  *node[V]
	write bool
}

func (txn *treeTxn[V]) get(k []byte) (V, bool) {
	return txn.root.get(k)
}

func (txn *treeTxn[V]) set(k []byte, v V) {
	if txn.write {
		txn.root = txn.root.set(k, v)
	}
}

func (txn *treeTxn[V]) del(k []byte) {
	if txn.write {
		txn.root = txn.root.del(k)
	}
}

func (txn *treeTxn[V]) commit() *tree[V] {
	return &tree[V]{root: txn.root}
}

func (txn *treeTxn[V]) cursor() *treeCursor[V] {
	return &treeCursor[V]{txn: txn}
}
