package memdb

type TableSelection[V any] struct {
	table Table[V]
	tx    *Txn
	ids   *treeTxn[struct{}]
	idx   *treeTxn[V]
	order *treeTxn[*tree[struct{}]]
	dir   OrderDirection
}

func (t *TableSelection[V]) pageUnorderedUnfilteredASC(limit, offset int) []V {
	out := []V{}
	at := 0
	c := t.idx.cursor()
	ok := c.first()
	for ok {
		if at >= offset {
			out = append(out, c.val())
		}
		if limit > 0 && len(out) >= limit {
			break
		}
		at++
		ok = c.next()
	}
	return out
}

func (t *TableSelection[V]) pageUnorderedUnfilteredDESC(limit, offset int) []V {
	out := []V{}
	at := 0
	c := t.idx.cursor()
	ok := c.last()
	for ok {
		if at >= offset {
			out = append(out, c.val())
		}
		if limit > 0 && len(out) >= limit {
			break
		}
		at++
		ok = c.prev()
	}
	return out
}

func (t *TableSelection[V]) pageOrderedUnfilteredASC(limit, offset int) []V {
	out := []V{}
	at := 0
	c := t.order.cursor()
	ok := c.first()
	for ok {
		cc := c.val().txn(false).cursor()
		okk := cc.first()
		for okk {
			if at >= offset {
				v, _ := t.idx.get(cc.key())
				out = append(out, v)
			}
			if limit > 0 && len(out) >= limit {
				break
			}
			at++
			okk = cc.next()
		}
		if limit > 0 && len(out) >= limit {
			break
		}
		ok = c.next()
	}
	return out
}

func (t *TableSelection[V]) pageOrderedUnfilteredDESC(limit, offset int) []V {
	out := []V{}
	at := 0
	c := t.order.cursor()
	ok := c.last()
	for ok {
		cc := c.val().txn(false).cursor()
		okk := cc.first()
		for okk {
			if at >= offset {
				v, _ := t.idx.get(cc.key())
				out = append(out, v)
			}
			if limit > 0 && len(out) >= limit {
				break
			}
			at++
			okk = cc.next()
		}
		if limit > 0 && len(out) >= limit {
			break
		}
		ok = c.prev()
	}
	return out
}

func (t *TableSelection[V]) pageUnorderedFilteredASC(limit, offset int) []V {
	out := []V{}
	at := 0
	c := t.ids.cursor()
	ok := c.first()
	for ok {
		if at >= offset {
			v, ok := t.idx.get(c.key())
			if ok {
				out = append(out, v)
			}
		}
		if limit > 0 && len(out) >= limit {
			break
		}
		ok = c.next()
		at++
	}
	return out
}

func (t *TableSelection[V]) pageUnorderedFilteredDESC(limit, offset int) []V {
	out := []V{}
	at := 0
	c := t.ids.cursor()
	ok := c.last()
	for ok {
		if at >= offset {
			v, ok := t.idx.get(c.key())
			if ok {
				out = append(out, v)
			}
		}
		if limit > 0 && len(out) >= limit {
			break
		}
		ok = c.prev()
		at++
	}
	return out
}

func (t *TableSelection[V]) pageOrderedFilteredASC(limit, offset int) []V {
	out := []V{}
	at := 0
	c := t.order.cursor()
	ok := c.first()
	for ok {
		cc := c.val().txn(false).cursor()
		okk := cc.first()
		for okk {
			if at >= offset {
				_, has := t.ids.get(cc.key())
				if has {
					v, _ := t.idx.get(cc.key())
					out = append(out, v)
				}
			}
			if limit > 0 && len(out) >= limit {
				break
			}
			at++
			okk = cc.next()
		}
		if limit > 0 && len(out) >= limit {
			break
		}
		ok = c.next()
	}
	return out
}

func (t *TableSelection[V]) pageOrderedFilteredDESC(limit, offset int) []V {
	out := []V{}
	at := 0
	c := t.order.cursor()
	ok := c.last()
	for ok {
		cc := c.val().txn(false).cursor()
		okk := cc.first()
		for okk {
			if at >= offset {
				_, has := t.ids.get(cc.key())
				if has {
					v, _ := t.idx.get(cc.key())
					out = append(out, v)
				}
			}
			if limit > 0 && len(out) >= limit {
				break
			}
			at++
			okk = cc.next()
		}
		if limit > 0 && len(out) >= limit {
			break
		}
		ok = c.prev()
	}
	return out
}

func (t *TableSelection[V]) count() int {
	res := 0
	if t.ids == nil {
		c := t.idx.cursor()
		ok := c.first()
		for ok {
			res++
			ok = c.next()
		}
	} else {
		c := t.ids.cursor()
		ok := c.first()
		for ok {
			res++
			ok = c.next()
		}
	}
	return res
}

func (t *TableSelection[V]) page(limit, offset int) []V {
	ordered := t.order != nil
	filtered := t.ids != nil
	asc := t.dir == Asc
	switch {
	case !ordered && !filtered && asc:
		return t.pageUnorderedUnfilteredASC(limit, offset)
	case !ordered && !filtered && !asc:
		return t.pageUnorderedUnfilteredDESC(limit, offset)
	case !ordered && filtered && asc:
		return t.pageUnorderedFilteredASC(limit, offset)
	case !ordered && filtered && !asc:
		return t.pageUnorderedFilteredDESC(limit, offset)
	case ordered && !filtered && asc:
		return t.pageOrderedUnfilteredASC(limit, offset)
	case ordered && !filtered && !asc:
		return t.pageOrderedUnfilteredDESC(limit, offset)
	case ordered && filtered && asc:
		return t.pageOrderedFilteredASC(limit, offset)
	case ordered && filtered && !asc:
		return t.pageOrderedFilteredDESC(limit, offset)
	}
	panic("unreachable")
}

func (t *TableSelection[V]) one() (V, error) {
	data := t.page(1, 0)
	if len(data) == 0 {
		return *new(V), ErrNotFound
	}
	return data[0], nil
}
