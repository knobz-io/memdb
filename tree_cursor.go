package memdb

type treeCursor[V any] struct {
	txn  *treeTxn[V]
	node *node[V]
}

func (c *treeCursor[V]) key() []byte {
	return c.node.k
}

func (c *treeCursor[V]) val() V {
	return c.node.v
}

func (c *treeCursor[V]) next() bool {
	if c.node == nil {
		c.node = c.txn.root.min()
		return c.node != nil
	}
	c.node = c.txn.root.successor(c.node.k)
	return c.node != nil
}

func (c *treeCursor[V]) prev() bool {
	if c.node == nil {
		c.node = c.txn.root.max()
		return c.node != nil
	}
	c.node = c.txn.root.predecessor(c.node.k)
	return c.node != nil
}

func (c *treeCursor[V]) seek(k []byte) bool {
	c.node = c.txn.root.seekNode(k)
	return c.node != nil
}

func (c *treeCursor[V]) first() bool {
	c.node = c.txn.root.min()
	return c.node != nil
}

func (c *treeCursor[V]) last() bool {
	c.node = c.txn.root.max()
	return c.node != nil
}
