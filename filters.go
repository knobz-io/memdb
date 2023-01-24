package memdb

import "bytes"

type IndexMap[V any] struct {
	arr []Field[V]
	m   map[Field[V]]int
	n   int
}

func (f IndexMap[V]) add(ff Field[V]) IndexMap[V] {
	f.arr = append(f.arr, ff)
	f.m[ff] = f.n
	f.n++
	return f
}

type Field[V any] interface {
	KeyOf(v V) Key
	field()
}

type StringField[V any] struct {
	fn func(v V) string
}

func (f *StringField[V]) Is(v string) *EqualCond[V] {
	return &EqualCond[V]{f, StringKey(v)}
}

func (f *StringField[V]) KeyOf(v V) Key {
	return StringKey(f.fn(v))
}

func (f *StringField[V]) LessThan(v string) *LessThanCond[V] {
	return &LessThanCond[V]{f, StringKey(v)}
}

func (f *StringField[V]) LessThanOrEqual(v string) *LessThanOrEqualCond[V] {
	return &LessThanOrEqualCond[V]{f, StringKey(v)}
}

func (f *StringField[V]) GreaterThan(v string) *GreaterThanCond[V] {
	return &GreaterThanCond[V]{f, StringKey(v)}
}

func (f *StringField[V]) GreaterThanOrEqual(v string) *GreaterThanOrEqualCond[V] {
	return &GreaterThanOrEqualCond[V]{f, StringKey(v)}
}

func (f *StringField[V]) field() {}

type IntField[V any] struct {
	fn func(v V) int
}

func (f *IntField[V]) KeyOf(v V) Key {
	return IntKey(f.fn(v))
}

func (f *IntField[V]) field() {}

func (f *IntField[V]) Is(v int) *EqualCond[V] {
	return &EqualCond[V]{f, IntKey(v)}
}

func (f *IntField[V]) IsLessThan(v int) *LessThanCond[V] {
	return &LessThanCond[V]{f, IntKey(v)}
}

func (f *IntField[V]) IsLessThanOrEqual(v int) *LessThanOrEqualCond[V] {
	return &LessThanOrEqualCond[V]{f, IntKey(v)}
}

func (f *IntField[V]) IsGreaterThan(v int) *GreaterThanCond[V] {
	return &GreaterThanCond[V]{f, IntKey(v)}
}

func (f *IntField[V]) IsGreaterThanOrEqual(v int) *GreaterThanOrEqualCond[V] {
	return &GreaterThanOrEqualCond[V]{f, IntKey(v)}
}

type FloatField[V any] struct {
	fn func(v V) float64
}

func (f *FloatField[V]) KeyOf(v V) Key {
	return FloatKey(f.fn(v))
}

func (f *FloatField[V]) Is(v float64) *EqualCond[V] {
	return &EqualCond[V]{f, FloatKey(v)}
}

func (f *FloatField[V]) LessThan(v float64) *LessThanCond[V] {
	return &LessThanCond[V]{f, FloatKey(v)}
}

func (f *FloatField[V]) LessThanOrEqual(v float64) *LessThanOrEqualCond[V] {
	return &LessThanOrEqualCond[V]{f, FloatKey(v)}
}

func (f *FloatField[V]) GreaterThan(v float64) *GreaterThanCond[V] {
	return &GreaterThanCond[V]{f, FloatKey(v)}
}

func (f *FloatField[V]) GreaterThanOrEqual(v float64) *GreaterThanOrEqualCond[V] {
	return &GreaterThanOrEqualCond[V]{f, FloatKey(v)}
}

func (f *FloatField[V]) field() {}

type BoolField[V any] struct {
	fn func(v V) bool
}

func (f *BoolField[V]) KeyOf(v V) Key {
	return BoolKey(f.fn(v))
}

func (f *BoolField[V]) IsTrue() *EqualCond[V] {
	return &EqualCond[V]{f, BoolKey(true)}
}

func (f *BoolField[V]) IsFalse() *EqualCond[V] {
	return &EqualCond[V]{f, BoolKey(false)}
}

func (f *BoolField[V]) field() {}

type BinaryField[V any] struct {
	fn func(v V) []byte
}

func (f *BinaryField[V]) KeyOf(v V) Key {
	return BinaryKey(f.fn(v))
}

func (f *BinaryField[V]) field() {}

func (f *BinaryField[V]) Is(v []byte) *EqualCond[V] {
	return &EqualCond[V]{f, BinaryKey(v)}
}

func (f *BinaryField[V]) LessThan(v []byte) *LessThanCond[V] {
	return &LessThanCond[V]{f, BinaryKey(v)}
}

func (f *BinaryField[V]) LessThanOrEqual(v []byte) *LessThanOrEqualCond[V] {
	return &LessThanOrEqualCond[V]{f, BinaryKey(v)}
}

func (f *BinaryField[V]) GreaterThan(v []byte) *GreaterThanCond[V] {
	return &GreaterThanCond[V]{f, BinaryKey(v)}
}

func (f *BinaryField[V]) GreaterThanOrEqual(v []byte) *GreaterThanOrEqualCond[V] {
	return &GreaterThanOrEqualCond[V]{f, BinaryKey(v)}
}

type MultiField[V any] struct {
	fn func(v V) MultiKey
}

func (f *MultiField[V]) KeyOf(v V) Key {
	return f.fn(v)
}

func (f *MultiField[V]) field() {}

func (f *MultiField[V]) Is(k MultiKey) *EqualCond[V] {
	return &EqualCond[V]{f, k}
}

type indexCond[V any] interface {
	field() Field[V]
	matches(k []byte) bool
	Matches(v V) bool
}

type EqualCond[V any] struct {
	f   Field[V]
	key Key
}

func (c *EqualCond[V]) field() Field[V] {
	return c.f
}

func (c *EqualCond[V]) matches(k []byte) bool {
	return bytes.Equal(k, c.key.Bytes())
}

func (c *EqualCond[V]) Matches(v V) bool {
	return c.matches(c.f.KeyOf(v).Bytes())
}

type LessThanCond[V any] struct {
	f   Field[V]
	key Key
}

func (c *LessThanCond[V]) field() Field[V] {
	return c.f
}

func (c *LessThanCond[V]) Matches(v V) bool {
	return c.matches(c.f.KeyOf(v).Bytes())
}

func (c *LessThanCond[V]) matches(k []byte) bool {
	return bytes.Compare(k, c.key.Bytes()) < 0
}

type LessThanOrEqualCond[V any] struct {
	f   Field[V]
	key Key
}

func (c *LessThanOrEqualCond[V]) field() Field[V] {
	return c.f
}

func (c *LessThanOrEqualCond[V]) matches(k []byte) bool {
	return bytes.Compare(k, c.key.Bytes()) <= 0
}

func (c *LessThanOrEqualCond[V]) Matches(v V) bool {
	return c.matches(c.f.KeyOf(v).Bytes())
}

type GreaterThanCond[V any] struct {
	f   Field[V]
	key Key
}

func (c *GreaterThanCond[V]) field() Field[V] {
	return c.f
}

func (c *GreaterThanCond[V]) Matches(v V) bool {
	return c.matches(c.f.KeyOf(v).Bytes())
}

func (c *GreaterThanCond[V]) matches(k []byte) bool {
	return bytes.Compare(k, c.key.Bytes()) > 0
}

type GreaterThanOrEqualCond[V any] struct {
	f   Field[V]
	key Key
}

func (c *GreaterThanOrEqualCond[V]) field() Field[V] {
	return c.f
}

func (c *GreaterThanOrEqualCond[V]) Matches(v V) bool {
	return c.matches(c.f.KeyOf(v).Bytes())
}

func (c *GreaterThanOrEqualCond[V]) matches(k []byte) bool {
	return bytes.Compare(k, c.key.Bytes()) >= 0
}
