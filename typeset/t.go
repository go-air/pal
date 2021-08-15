// Copyright 2021 The pal authors (see AUTHORS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package typeset

type T struct {
	nodes []node
	hash  []Type
}

const (
	initCap = 128
)

func New() *T {
	res := &T{}
	res.nodes = make([]node, _endType, initCap)
	copy(res.nodes, basicNodes)
	res.hash = make([]Type, initCap)
	for i := Type(1); i < _endType; i++ {
		node := &res.nodes[i]
		node.hash = res.hashCode(i)
		ci := node.hash % initCap
		node.next = res.hash[ci]
		res.hash[ci] = i
	}
	return res
}

func (t *T) getSlice(elt Type) Type {
	ty, node := t.newNode()
	node.kind = Slice
	node.elem = elt
	node.hash = t.hashCode(ty)
	ci := node.hash % uint32(cap(t.hash))
	for {
		_ = ci

	}
	return ty
}

func (t *T) getPointer(elt Type) Type {
	return NoType
}

func (t *T) getChan(elt Type) Type {
	return NoType
}

func (t *T) getArray(elt Type, n int64) Type {
	return NoType
}

func (t *T) newNode() (Type, *node) {
	n := len(t.nodes)
	if n == cap(t.nodes) {
		t.grow()
	}
	t.nodes = t.nodes[:n+1]
	node := &t.nodes[n]
	node.zero()
	return Type(n), node
}

func (t *T) grow() {
	ncap := uint32(cap(t.nodes) * 2)
	tnodes := make([]node, len(t.nodes), ncap)
	thash := make([]uint32, ncap)
	copy(tnodes, t.nodes)
	for i := range t.nodes {
		ty := uint32(i)
		node := &t.nodes[i]
		ci := node.hash % ncap
		node.next = Type(thash[ci])
		thash[ci] = ty
	}
	t.nodes = t.nodes
	t.hash = thash
}

func (t *T) Equals(a, b Type) bool {
	return a == b
}

func (t *T) equals(a, b Type) bool {
	anode, bnode := &t.nodes[a], &t.nodes[b]
	if anode.kind != bnode.kind {
		return false
	}
	if anode.kind == Basic {
		return a == b
	}
	if anode.lsize != bnode.lsize {
		return false
	}
	switch anode.kind {
	case Pointer, Slice, Chan:
		return t.equals(anode.elem, bnode.elem)
	case Array:
		return t.equals(anode.elem, bnode.elem)
	case Struct, Interface: // interface methods are sorted
		return t.namedsEqual(anode.fields, bnode.fields)
	case Map:
		return t.equals(anode.elem, bnode.elem) && t.equals(anode.key, bnode.key)
	case Func:
		return t.namedsEqual(anode.params, bnode.params) && t.namedsEqual(anode.results, bnode.results)
	default:
		panic("bad kind")
	}
	return true
}

func (t *T) namedsEqual(as, bs []named) bool {
	if len(as) != len(bs) {
		return false
	}
	for i := range as {
		anamed := as[i]
		bnamed := bs[i]
		if anamed.name != bnamed.name {
			return false
		}
		if !t.equals(anamed.typ, bnamed.typ) {
			return false
		}
	}
	return true
}

func (t *T) hashCode(ty Type) uint32 {
	result := uint32(ty)
	node := &t.nodes[ty]
	result *= uint32(node.kind) << 3
	if node.elem != NoType {
		result *= uint32(node.elem) << 5
	}
	if node.key != NoType {
		result *= uint32(node.key) << 7
	}
	result *= uint32(node.lsize) << 11
	result = result ^ hashNamed(node.fields)
	result = result ^ hashNamed(node.params)
	result = result ^ hashNamed(node.results)
	return result
}

func hashNamed(named []named) uint32 {
	result := uint32(1<<32 - 1)
	for _, nd := range named {
		result <<= 13
		result *= uint32(nd.typ)
		n := len(nd.name)
		for i := 0; i < n; i++ {
			result <<= 7
			result *= uint32(nd.name[i])
		}
	}
	return result
}
