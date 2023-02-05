package memdb

import (
	"bytes"
)

type TableLister[V any] struct {
	table Table[V]
	tx    *Txn
	conds []Cond[V]
	order Index[V]
	dir   OrderDirection
}

func (t *TableLister[V]) OrderBy(order Index[V]) *TableLister[V] {
	t.order = order
	return t
}

func (t *TableLister[V]) Asc() *TableLister[V] {
	t.dir = Asc
	return t
}

func (t *TableLister[V]) Desc() *TableLister[V] {
	t.dir = Desc
	return t
}

func (t *TableLister[V]) Where(conds ...Cond[V]) *TableLister[V] {
	t.conds = append(t.conds, conds...)
	return t
}

func (t *TableLister[V]) Count() (int, error) {
	return t.selector().count(), nil
}

func (t *TableLister[V]) Page(limit, offset int) ([]V, error) {
	selector := t.selector()
	return selector.page(limit, offset), nil
}

func (t *TableLister[V]) All() ([]V, error) {
	selector := t.selector()
	return selector.page(0, 0), nil
}

func (t *TableLister[V]) One() (V, error) {
	selector := t.selector()
	return selector.one()
}

func (t *TableLister[V]) Cursor() (*TableCursor[V], error) {
	return nil, nil
}

func (t *TableLister[V]) selector() *TableSelection[V] {
	indexed := []indexCond[V]{}
	basic := []Cond[V]{}
	for _, cond := range t.conds {
		if ic, ok := cond.(indexCond[V]); ok {
			indexed = append(indexed, ic)
		} else {
			basic = append(basic, cond)
		}
	}
	var ids *treeTxn[struct{}]
	if len(indexed) > 0 {
		idTree := makeTree[struct{}]()
		for i, cnd := range indexed {
			tmp := makeTree[struct{}]()
			f := cnd.field()
			idxID := t.table.idxm.m[f] + 1
			idx := (*treeTxn[*tree[struct{}]])(t.tx.tm[t.table.ref][uint8(idxID)])

			switch cnd := cnd.(type) {
			case *EqualCond[V]:
				subidx, ok := idx.get(cnd.key.Bytes())
				if ok {
					tmp = subidx
				}
			case *LessThanCond[V]:
				c := idx.cursor()
				k := cnd.key.Bytes()
				ok := c.seek(k)
				for ok && bytes.Equal(c.key(), k) { // skip maching
					ok = c.prev()
				}
				for ok {
					tmp = tmp.union(c.val())
					ok = c.prev()
				}
			case *LessThanOrEqualCond[V]:
				c := idx.cursor()
				k := cnd.key.Bytes()
				ok := c.seek(k)
				for ok {
					tmp = tmp.union(c.val())
					ok = c.prev()
				}
			case *GreaterThanCond[V]:
				c := idx.cursor()
				k := cnd.key.Bytes()
				ok := c.seek(k)
				for ok && bytes.Equal(c.key(), k) { // skip maching
					ok = c.next()
				}
				for ok {
					tmp = tmp.union(c.val())
					ok = c.next()
				}
			case *GreaterThanOrEqualCond[V]:
				c := idx.cursor()
				k := cnd.key.Bytes()
				ok := c.seek(k)
				for ok {
					tmp = tmp.union(c.val())
					ok = c.next()
				}
			}
			if i == 0 {
				idTree = tmp
			} else {
				idTree = idTree.intersectOptimized(tmp)
			}
		}
		ids = idTree.txn(false)
	}
	var order *treeTxn[*tree[struct{}]]
	if t.order != nil {
		order = (*treeTxn[*tree[struct{}]])(t.tx.tm[t.table.ref][uint8(t.table.idxm.m[t.order]+1)])
	}
	selection := (*treeTxn[V])(t.tx.tm[t.table.ref][0])
	return &TableSelection[V]{table: t.table, tx: t.tx, idx: selection, ids: ids, order: order, dir: t.dir}
}
