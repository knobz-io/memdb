package memdb

import (
	"bytes"
	"encoding/binary"
)

type KeyFunc[V any] func(V) Key

type Key interface {
	Bytes() []byte
}

type StringKey string

func (s StringKey) Bytes() []byte {
	return []byte(s)
}

type IntKey int

func (i IntKey) Bytes() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

type FloatKey float64

func (f FloatKey) Bytes() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(f))
	return buf
}

type BoolKey bool

func (b BoolKey) Bytes() []byte {
	if b {
		return []byte{1}
	}
	return []byte{0}
}

type BinaryKey []byte

func (b BinaryKey) Bytes() []byte {
	return b
}

type MultiKey []Key

func (mk MultiKey) Bytes() []byte {
	dst := [][]byte{}
	for _, k := range mk {
		dst = append(dst, k.Bytes())
	}
	return bytes.Join(dst, nil)
}
