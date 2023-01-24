package memdb

import (
	"sync/atomic"
	"unsafe"
)

type callbacks[V any] struct {
	setfn []func(tx *Txn, v V)
	updfn []func(tx *Txn, v V, prev V)
	delfn []func(tx *Txn, v V)
}

type DB struct {
	tm       map[interface{}]map[uint8]*unsafe.Pointer
	txfn     map[interface{}]map[uint8]func(unsafe.Pointer) unsafe.Pointer
	commitfn map[interface{}]map[uint8]func(unsafe.Pointer) unsafe.Pointer
	indexm   map[interface{}]int
}

func Init(tables ...TableType) (*DB, error) {
	db := &DB{
		// root:     unsafe.Pointer(iradix.New[*iradix.Tree[[]byte]]()),
		tm:       map[interface{}]map[uint8]*unsafe.Pointer{},
		txfn:     map[interface{}]map[uint8]func(unsafe.Pointer) unsafe.Pointer{},
		commitfn: map[interface{}]map[uint8]func(unsafe.Pointer) unsafe.Pointer{},
		indexm:   map[interface{}]int{},
	}
	for _, table := range tables {
		err := table.registerTable(db)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}

func (db *DB) Tx(write bool) *Txn {
	tm := make(map[interface{}]map[uint8]unsafe.Pointer, len(db.tm))
	for ref, n := range db.indexm {
		v := n
		tm[ref] = make(map[uint8]unsafe.Pointer, v)
		for j := 0; j <= v; j++ {
			p := db.tm[ref][uint8(j)]
			tm[ref][uint8(j)] = db.txfn[ref][uint8(j)](atomic.LoadPointer(p))
		}
	}
	return &Txn{
		// root:  root.Txn(),
		db:    db,
		tm:    tm,
		write: write,
	}
}

type Txn struct {
	write bool
	db    *DB
	tm    map[interface{}]map[uint8]unsafe.Pointer
}

func (tx *Txn) Commit() {
	if !tx.write {
		return
	}
	if tx.tm == nil /** || tx.root == nil */ {
		return // already committed
	}
	for ref, v := range tx.tm {
		for j, p := range v {
			atomic.StorePointer(tx.db.tm[ref][j], tx.db.commitfn[ref][j](p))
		}
	}
	tx.tm = nil
}

func (tx *Txn) Abort() {
	if !tx.write {
		return
	}
	if tx.tm == nil {
		return
	}
	tx.tm = nil
}

func (db *DB) WriteTx() *Txn {
	return db.Tx(true)
}

func (db *DB) ReadTx() *Txn {
	return db.Tx(false)
}
