package memdb

import (
	"reflect"
	"testing"
)

func Test_node_copy(t *testing.T) {
	type fields struct {
		bf    int
		h     int
		k     []byte
		v     int
		left  *node[int]
		right *node[int]
	}
	tests := []struct {
		name   string
		fields *fields
		want   *node[int]
	}{
		{
			name: "nil",
			want: nil,
		},
		{
			name: "empty",
			fields: &fields{
				bf:    0,
				h:     0,
				k:     nil,
				v:     0,
				left:  nil,
				right: nil,
			},
			want: &node[int]{
				bf:    0,
				h:     0,
				k:     nil,
				v:     0,
				left:  nil,
				right: nil,
			},
		},
		{
			name: "non_empty",
			fields: &fields{
				bf:    1,
				h:     2,
				k:     []byte("a"),
				v:     3,
				left:  &node[int]{},
				right: &node[int]{},
			},
			want: &node[int]{
				bf:    1,
				h:     2,
				k:     []byte("a"),
				v:     3,
				left:  &node[int]{},
				right: &node[int]{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n *node[int]
			if tt.fields != nil {
				n = &node[int]{
					bf:    tt.fields.bf,
					h:     tt.fields.h,
					k:     tt.fields.k,
					v:     tt.fields.v,
					left:  tt.fields.left,
					right: tt.fields.right,
				}
			}
			if got := n.copy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("node.copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_node_rotateLeft(t *testing.T) {
	tests := []struct {
		name string
		node *node[int]
		want *node[int]
	}{
		{
			name: "nil",
			want: nil,
		},
		{
			name: "empty",
			node: &node[int]{},
			want: &node[int]{},
		},
		{
			name: "one",
			node: &node[int]{
				bf: 1,
				h:  2,
				k:  []byte("a"),
				v:  3,
			},
			want: &node[int]{
				bf: 1,
				h:  2,
				k:  []byte("a"),
				v:  3,
			},
		},
		{
			name: "two",
			node: &node[int]{
				bf: 1,
				h:  2,
				k:  []byte("a"),
				v:  3,
				right: &node[int]{
					bf: 1,
					h:  2,
					k:  []byte("b"),
					v:  4,
				},
			},
			want: &node[int]{
				bf: -1,
				h:  2,
				k:  []byte("b"),
				v:  4,
				left: &node[int]{
					bf: 0,
					h:  1,
					k:  []byte("a"),
					v:  3,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.rotateLeft(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("node.rotateLeft() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_node_rotateRight(t *testing.T) {
	tests := []struct {
		name string
		node *node[int]
		want *node[int]
	}{
		{
			name: "nil",
			want: nil,
		},
		{
			name: "empty",
			node: &node[int]{},
			want: &node[int]{},
		},
		{
			name: "one",
			node: &node[int]{
				bf: 1,
				h:  2,
				k:  []byte("a"),
				v:  3,
			},
			want: &node[int]{
				bf: 1,
				h:  2,
				k:  []byte("a"),
				v:  3,
			},
		},
		{
			name: "two",
			node: &node[int]{
				bf: 1,
				h:  2,
				k:  []byte("a"),
				v:  3,
				left: &node[int]{
					bf: 1,
					h:  2,
					k:  []byte("b"),
					v:  4,
				},
			},
			want: &node[int]{
				bf: 1,
				h:  2,
				k:  []byte("b"),
				v:  4,
				right: &node[int]{
					bf: 0,
					h:  1,
					k:  []byte("a"),
					v:  3,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.rotateRight(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("node.rotateRight() = %v, want %v", got, tt.want)
			}
		})
	}
}
