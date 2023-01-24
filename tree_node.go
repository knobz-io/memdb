package memdb

import (
	"bytes"
)

type node[V any] struct {
	bf    int
	h     int
	k     []byte
	v     V
	left  *node[V]
	right *node[V]
}

func (n *node[V]) copy() *node[V] {
	if n == nil {
		return nil
	}
	return &node[V]{
		bf:    n.bf,
		h:     n.h,
		k:     n.k,
		v:     n.v,
		left:  n.left,
		right: n.right,
	}
}

func (n *node[V]) rotateLeft() *node[V] {
	if n == nil || n.right == nil {
		return n
	}
	r := n.right
	n = n.copy()
	n.right = r.left
	r = r.copy()
	r.left = n
	n.updateHeight()
	r.updateHeight()
	return r
}

func (n *node[V]) rotateRight() *node[V] {
	if n == nil || n.left == nil {
		return n
	}
	l := n.left
	n = n.copy()
	n.left = l.right
	l = l.copy()
	l.right = n
	n.updateHeight()
	l.updateHeight()
	return l
}

func (n *node[V]) updateHeight() {
	if n == nil {
		return
	}
	n.h = 1 + max(n.left.height(), n.right.height())
	n.bf = n.right.height() - n.left.height()
}

func (n *node[V]) height() int {
	if n == nil {
		return 0
	}
	return n.h
}

func (n *node[V]) balance() *node[V] {
	if n == nil {
		return nil
	}
	n = n.copy()
	if n.bf == -2 {
		if n.left.bf <= 0 {
			n = n.rotateRight()
		} else {
			n.left = n.left.rotateLeft()
			n = n.rotateRight()
		}
	} else if n.bf == 2 {
		if n.right.bf >= 0 {
			n = n.rotateLeft()
		} else {
			n.right = n.right.rotateRight()
			n = n.rotateLeft()
		}
	}
	return n
}

func (n *node[V]) set(k []byte, v V) *node[V] {
	if n == nil {
		return &node[V]{k: k, v: v}
	}
	n = n.copy()
	cmp := bytes.Compare(k, n.k)
	if cmp == 0 {
		n.v = v
	} else if cmp < 0 {
		n.left = n.left.set(k, v)
	} else {
		n.right = n.right.set(k, v)
	}
	n.updateHeight()
	return n.balance()
}

func (n *node[V]) get(k []byte) (V, bool) {
	if n == nil {
		return *new(V), false
	}
	cmp := bytes.Compare(k, n.k)
	if cmp == 0 {
		return n.v, true
	} else if cmp < 0 {
		return n.left.get(k)
	} else {
		return n.right.get(k)
	}
}

func (n *node[V]) min() *node[V] {
	if n == nil || n.left == nil {
		return n
	}
	return n.left.min()
}

func (n *node[V]) max() *node[V] {
	if n == nil || n.right == nil {
		return n
	}
	return n.right.max()
}

func (n *node[V]) delLeft() *node[V] {
	if n == nil {
		return nil
	}
	if n.left == nil {
		return n.right
	}
	n = n.copy()
	n.left = n.left.delLeft()
	n.updateHeight()
	return n.balance()
}

func (n *node[V]) del(k []byte) *node[V] {
	if n == nil {
		return nil
	}
	n = n.copy()
	cmp := bytes.Compare(k, n.k)
	if cmp < 0 {
		n.left = n.left.del(k)
	} else if cmp > 0 {
		n.right = n.right.del(k)
	} else {
		if n.left == nil {
			return n.right
		}
		if n.right == nil {
			return n.left
		}
		m := n.right
		for m.left != nil {
			m = m.left
		}
		n.k = m.k
		n.v = m.v
		n.right = n.right.delLeft()
	}
	n.updateHeight()
	return n.balance()
}

func (n *node[V]) successor(k []byte) *node[V] {
	if n == nil {
		return nil
	}
	cmp := bytes.Compare(k, n.k)
	if cmp == 0 {
		return n.right.min()
	} else if cmp < 0 {
		s := n.left.successor(k)
		if s == nil {
			return n
		}
		return s
	} else {
		return n.right.successor(k)
	}
}

func (n *node[V]) predecessor(k []byte) *node[V] {
	if n == nil {
		return nil
	}
	cmp := bytes.Compare(k, n.k)
	if cmp == 0 {
		return n.left.max()
	} else if cmp < 0 {
		return n.left.predecessor(k)
	} else {
		p := n.right.predecessor(k)
		if p == nil {
			return n
		}
		return p
	}
}

func (n *node[V]) seekNode(k []byte) *node[V] {
	if n == nil {
		return nil
	}
	cmp := bytes.Compare(k, n.k)
	if cmp == 0 {
		return n
	} else if cmp < 0 {
		nn := n.left.seekNode(k)
		if nn == nil {
			return n
		} else {
			return nn
		}
	} else {
		nn := n.right.seekNode(k)
		if nn == nil {
			return n
		} else {
			return nn
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
