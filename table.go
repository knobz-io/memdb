package memdb

import (
	"bytes"
	"errors"
	"sync/atomic"
	"unsafe"
)

type TableType interface {
	registerTable(db *DB) error
	table()
}

type Table[V any] struct {
	ref  *V
	fn   KeyFunc[V]
	idxm IndexMap[V]
	cb   callbacks[V]
}

func NewTable[V any](fn KeyFunc[V]) Table[V] {
	return Table[V]{
		ref: new(V),
		fn:  fn,
		idxm: IndexMap[V]{
			m: make(map[Field[V]]int),
		},
	}
}

func (t Table[V]) IndexString(fn func(V) string) (Table[V], *StringIndex[V]) {
	f := &StringIndex[V]{fn}
	t.idxm = t.idxm.add(f)
	t = t.registerIndex(f)
	return t, f
}

func (t Table[V]) IndexInt(fn func(V) int) (Table[V], *IntIndex[V]) {
	f := &IntIndex[V]{fn}
	t.idxm = t.idxm.add(f)
	t = t.registerIndex(f)
	return t, f
}

func (t Table[V]) IndexFloat(fn func(V) float64) (Table[V], *FloatIndex[V]) {
	f := &FloatIndex[V]{fn}
	t.idxm = t.idxm.add(f)
	t = t.registerIndex(f)
	return t, f
}

func (t Table[V]) IndexBool(fn func(V) bool) (Table[V], *BoolIndex[V]) {
	f := &BoolIndex[V]{fn}
	t.idxm = t.idxm.add(f)
	t = t.registerIndex(f)
	return t, f
}

func (t Table[V]) IndexBinary(fn func(V) []byte) (Table[V], *BinaryIndex[V]) {
	f := &BinaryIndex[V]{fn}
	t.idxm = t.idxm.add(f)
	t = t.registerIndex(f)
	return t, f
}

func (t Table[V]) IndexMultiple(fn func(V) CombinedKey) (Table[V], *CombinedIndex[V]) {
	f := &CombinedIndex[V]{fn}
	t.idxm = t.idxm.add(f)
	t = t.registerIndex(f)
	return t, f
}

func (t Table[V]) data(tx *Txn) (*treeTxn[V], error) {
	p, ok := tx.tm[t.ref][0]
	if !ok {
		return nil, errors.New("memdb: table not found in transaction")
	}
	return (*treeTxn[V])(p), nil
}

func (t Table[V]) Get(tx *Txn, id Key) (V, error) {
	data, err := t.data(tx)
	if err != nil {
		return *new(V), err
	}
	v, ok := data.get(id.Bytes())
	if !ok {
		return *new(V), ErrNotFound
	}
	return v, nil
}

func (t Table[V]) Set(tx *Txn, v V) error {
	data, err := t.data(tx)
	if err != nil {
		return err
	}
	k := t.fn(v).Bytes()
	prev, ok := data.get(k)

	data.set(k, v)
	if ok {
		for _, fn := range t.cb.updfn {
			fn(tx, v, prev)
		}
	} else {
		for _, fn := range t.cb.setfn {
			fn(tx, v)
		}
	}
	return nil
}

func (t Table[V]) SetMulti(tx *Txn, vs []V) error {
	data, err := t.data(tx)
	if err != nil {
		return err
	}
	for _, v := range vs {
		k := t.fn(v).Bytes()
		prev, ok := data.get(k)

		data.set(k, v)
		if ok {
			for _, fn := range t.cb.updfn {
				fn(tx, v, prev)
			}
		} else {
			for _, fn := range t.cb.setfn {
				fn(tx, v)
			}
		}
	}
	return nil
}

func (t Table[V]) DelMulti(tx *Txn, pks []Key) error {
	data, err := t.data(tx)
	if err != nil {
		return err
	}
	for _, pk := range pks {
		k := pk.Bytes()
		v, ok := data.get(k)
		if !ok {
			continue
		}
		data.del(k)
		for _, fn := range t.cb.delfn {
			fn(tx, v)
		}
	}
	return nil
}

func (t Table[V]) Del(tx *Txn, pk Key) error {
	data, err := t.data(tx)
	if err != nil {
		return err
	}
	k := pk.Bytes()
	v, ok := data.get(k)
	if !ok {
		return nil
	}
	data.del(k)
	for _, fn := range t.cb.delfn {
		fn(tx, v)
	}
	return nil
}

func (t Table[V]) Select(tx *Txn) *TableLister[V] {
	return &TableLister[V]{
		table: t,
		tx:    tx,
	}
}

func (t Table[V]) table() {}

func (t Table[V]) registerTable(db *DB) error {
	if t.ref == nil {
		return errors.New("memdb: table is not referenced")
	}
	if _, ok := db.tm[t.ref]; ok {
		return errors.New("memdb: table already registered")
	}
	root := unsafe.Pointer(makeTree[V]())
	if t.idxm.n > 255 {
		return errors.New("memdb: too many indexes")
	}
	n := t.idxm.n
	db.indexm[t.ref] = t.idxm.n
	db.tm[t.ref] = make(map[uint8]*unsafe.Pointer, n)
	db.txfn[t.ref] = make(map[uint8]func(unsafe.Pointer) unsafe.Pointer, n)
	db.commitfn[t.ref] = make(map[uint8]func(unsafe.Pointer) unsafe.Pointer, n)
	// set table root index
	db.tm[t.ref][0] = &root
	db.txfn[t.ref][0] = func(p unsafe.Pointer) unsafe.Pointer {
		idx := (*tree[V])(p)
		return unsafe.Pointer(idx.txn(true))
	}
	db.commitfn[t.ref][0] = func(txp unsafe.Pointer) unsafe.Pointer {
		tx := (*treeTxn[V])(txp)
		return unsafe.Pointer(tx.commit())
	}
	// set filter indexes for each combination
	for i := 1; i <= n; i++ {
		idx := unsafe.Pointer(makeTree[*tree[struct{}]]())
		db.tm[t.ref][uint8(i)] = &idx
		db.txfn[t.ref][uint8(i)] = func(p unsafe.Pointer) unsafe.Pointer {
			idx := (*tree[*tree[struct{}]])(atomic.LoadPointer(&p))
			return unsafe.Pointer(idx.txn(true))
		}
		db.commitfn[t.ref][uint8(i)] = func(txp unsafe.Pointer) unsafe.Pointer {
			tx := (*treeTxn[*tree[struct{}]])(txp)
			return unsafe.Pointer(tx.commit())
		}
	}
	return nil
}

func (t Table[V]) registerIndex(f Field[V]) Table[V] {
	i := t.idxm.m[f]
	t.cb.setfn = append(t.cb.setfn, func(tx *Txn, v V) {
		idx := (*treeTxn[*tree[struct{}]])(tx.tm[t.ref][uint8(i+1)])
		id := t.fn(v)
		k := f.KeyOf(v)
		subidx, ok := idx.get(k.Bytes())
		if !ok {
			subidx = makeTree[struct{}]()
		}
		subtx := subidx.txn(true)
		subtx.set(id.Bytes(), struct{}{})
		idx.set(k.Bytes(), subtx.commit())
	})
	t.cb.delfn = append(t.cb.delfn, func(tx *Txn, v V) {
		idx := (*treeTxn[*tree[struct{}]])(tx.tm[t.ref][uint8(i+1)])
		id := t.fn(v)
		k := f.KeyOf(v)
		subidx, ok := idx.get(k.Bytes())
		if !ok {
			return
		}
		subtx := subidx.txn(true)
		subtx.del(id.Bytes())
		if subtx.root.height() == 0 {
			idx.del(k.Bytes())
		} else {
			idx.set(k.Bytes(), subtx.commit())
		}
	})
	t.cb.updfn = append(t.cb.updfn, func(tx *Txn, v, prev V) {
		newk := f.KeyOf(v)
		prevk := f.KeyOf(prev)
		if bytes.Equal(newk.Bytes(), prevk.Bytes()) {
			return
		}
		idx := (*treeTxn[*tree[struct{}]])(tx.tm[t.ref][uint8(i+1)])
		id := t.fn(v)
		prevSubidx, ok := idx.get(prevk.Bytes())
		if ok {
			prevSubtx := prevSubidx.txn(true)
			prevSubtx.del(id.Bytes())
			if prevSubtx.root.height() == 0 {
				idx.del(prevk.Bytes())
			} else {
				idx.set(prevk.Bytes(), prevSubtx.commit())
			}
		}
		subidx, ok := idx.get(newk.Bytes())
		if !ok {
			subidx = makeTree[struct{}]()
		}
		subtx := subidx.txn(true)
		subtx.set(id.Bytes(), struct{}{})
		idx.set(newk.Bytes(), subtx.commit())
	})
	return t
}
