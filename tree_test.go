package memdb

import (
	"reflect"
	"testing"
)

type testTreeBuilder[V any] struct {
	t *tree[V]
}

func (b testTreeBuilder[V]) add(key string, val V) testTreeBuilder[V] {
	tx := b.t.txn(true)
	tx.set([]byte(key), val)
	return testTreeBuilder[V]{tx.commit()}
}

func (b testTreeBuilder[V]) finalize() *tree[V] {
	return b.t
}

func makeTestTree[V any]() *testTreeBuilder[V] {
	return &testTreeBuilder[V]{
		t: makeTree[V](),
	}
}

func makeTestTreeMap[V any](t *tree[V]) map[string]V {
	m := make(map[string]V)
	c := t.txn(false).cursor()
	ok := c.first()
	for ok {
		m[string(c.key())] = c.val()
		ok = c.next()
	}
	return m
}

func Test_tree_union(t *testing.T) {
	type args struct {
		o *tree[int]
	}
	tests := []struct {
		name string
		t    *tree[int]
		args args
		want *tree[int]
	}{
		{
			name: "empty_both",
			t:    makeTree[int](),
			args: args{o: makeTree[int]()},
			want: makeTree[int](),
		},
		{
			name: "empty_left",
			t:    makeTree[int](),
			args: args{o: makeTestTree[int]().add("a", 1).finalize()},
			want: makeTestTree[int]().add("a", 1).finalize(),
		},
		{
			name: "empty_right",
			t:    makeTestTree[int]().add("a", 1).finalize(),
			args: args{o: makeTree[int]()},
			want: makeTestTree[int]().add("a", 1).finalize(),
		},
		{
			name: "one_element_each",
			t:    makeTestTree[int]().add("a", 1).finalize(),
			args: args{o: makeTestTree[int]().add("b", 2).finalize()},
			want: makeTestTree[int]().add("a", 1).add("b", 2).finalize(),
		},
		{
			name: "one_element_each_same_key",
			t:    makeTestTree[int]().add("a", 1).finalize(),
			args: args{o: makeTestTree[int]().add("a", 2).finalize()},
			want: makeTestTree[int]().add("a", 2).finalize(),
		},
		{
			name: "two_elements_each_different_keys",
			t:    makeTestTree[int]().add("a", 1).add("b", 2).finalize(),
			args: args{o: makeTestTree[int]().add("c", 3).add("d", 4).finalize()},
			want: makeTestTree[int]().add("a", 1).add("b", 2).add("c", 3).add("d", 4).finalize(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeTestTreeMap(tt.t.union(tt.args.o)); !reflect.DeepEqual(got, makeTestTreeMap(tt.want)) {
				t.Errorf("tree.union() = %v, want %v", got, tt.want)
			}
		})
	}
}
