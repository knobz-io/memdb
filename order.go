package memdb

type OrderDirection int

const (
	Asc OrderDirection = iota
	Desc
)

type OrderRule[V any] struct {
	index Index[V]
	dir   OrderDirection
}
