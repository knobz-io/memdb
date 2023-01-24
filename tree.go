package memdb

import (
	"bytes"
)

type tree[V any] struct {
	root *node[V]
}

func (indx *tree[V]) txn(write bool) *treeTxn[V] {
	return &treeTxn[V]{root: indx.root, write: write}
}

func makeTree[V any]() *tree[V] {
	return &tree[V]{root: nil}
}

// intersect returns the intersection of two sorted slices.
func (idx *tree[V]) intersect(o *tree[V]) *tree[V] {
	if idx.root == nil || o.root == nil {
		return makeTree[V]()
	}
	dst := makeTree[V]().txn(true)
	c1 := idx.txn(false).cursor()
	c2 := o.txn(false).cursor()
	// use traversal over sorted trees to find the intersection
	ok1 := c1.first()
	ok2 := c2.first()
	for ok1 && ok2 {
		cmp := bytes.Compare(c1.key(), c2.key())
		if cmp == 0 {
			dst.set(c1.key(), c1.val())
			ok1 = c1.next()
			ok2 = c2.next()
		} else if cmp < 0 {
			ok1 = c1.next()
		} else {
			ok2 = c2.next()
		}
	}
	return dst.commit()
}

// intersectOptimized returns the intersection of two sorted slices.
func (idx *tree[V]) intersectOptimized(o *tree[V]) *tree[V] {
	if idx.root == nil || o.root == nil {
		return makeTree[V]()
	}
	dst := makeTree[V]().txn(true)
	c1 := idx.txn(false).cursor()
	c2 := o.txn(false).cursor()
	// use traversal over sorted trees to find the intersection
	ok1 := c1.first()
	ok2 := c2.first()
	for ok1 && ok2 {
		cmp := bytes.Compare(c1.key(), c2.key())
		if cmp == 0 {
			dst.set(c1.key(), c1.val())
			ok1 = c1.next()
			ok2 = c2.next()
		} else if cmp < 0 {
			// if the current key in c1 is less than the current key in c2
			// then we can skip all the keys in c1 that are less than the
			// current key in c2
			low := c1.key()
			ok1 = c1.seek(c2.key())
			if !ok1 {
				// if we can't seek to the current key in c2, then we need
				// to seek to the next key in c2
				ok2 = c2.next()
				if ok2 {
					ok1 = c1.seek(c2.key())
				}
			}
			if ok1 && bytes.Compare(c1.key(), low) <= 0 {
				// if the key we seeked to is less than or equal to the
				// low key we skipped, then we need to advance to the next
				// key
				ok1 = c1.next()
			}
		} else {
			// if the current key in c2 is less than the current key in c1
			// then we can skip all the keys in c2 that are less than the
			// current key in c1
			low := c2.key()
			ok2 = c2.seek(c1.key())
			if !ok2 {
				// if we can't seek to the current key in c1, then we need
				// to seek to the next key in c1
				ok1 = c1.next()
				if ok1 {
					ok2 = c2.seek(c1.key())
				}
			}
			if ok2 && bytes.Compare(c2.key(), low) <= 0 {
				// if the key we seeked to is less than or equal to the
				// low key we skipped, then we need to advance to the next
				// key
				ok2 = c2.next()
			}
		}
	}
	return dst.commit()
}

func (t *tree[V]) union(o *tree[V]) *tree[V] {
	if t.root == nil {
		return o
	}
	if o.root == nil {
		return t
	}
	dst := t.txn(true)
	c := o.txn(false).cursor()
	ok := c.first()
	for ok {
		dst.set(c.key(), c.val())
		ok = c.next()
	}
	return dst.commit()
}
