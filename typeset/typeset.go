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

import (
	"sort"
)

type TypeSet struct {
	nodes []node
	hash  []Type
}

const (
	initCap = 64
)

func New() *TypeSet {
	res := &TypeSet{}
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

func (t *TypeSet) Kind(ty Type) Kind {
	return t.nodes[ty].kind
}

func (t *TypeSet) IsObject(ty Type) bool {
	k := t.Kind(ty)
	return k != Basic && k != Tuple
}

func (t *TypeSet) Lsize(ty Type) int {
	return t.nodes[ty].lsize
}

func (t *TypeSet) Elem(ty Type) Type {
	return t.nodes[ty].elem
}

func (t *TypeSet) Key(ty Type) Type {
	return t.nodes[ty].key
}

func (t *TypeSet) ArrayLen(ty Type) int {
	node := &t.nodes[ty]
	n := node.lsize - 1
	eltSize := t.nodes[node.elem].lsize
	// nb lsize is never 0.
	if n%eltSize != 0 {
		panic("bad array len")
	}
	return n / eltSize
}

func (t *TypeSet) NumFields(ty Type) int {
	return len(t.nodes[ty].fields)
}

func (t *TypeSet) Field(ty Type, i int) (name string, fty Type, loff int) {
	f := t.nodes[ty].fields[i]
	return f.name, f.typ, f.loff
}

func (t *TypeSet) Recv(ty Type) Type {
	return t.nodes[ty].key
}

func (t *TypeSet) Variadic(ty Type) bool {
	return t.nodes[ty].variadic
}

func (t *TypeSet) NumParams(ty Type) int {
	return len(t.nodes[ty].params)
}

func (t *TypeSet) Param(ty Type, i int) (name string, pty Type) {
	param := t.nodes[ty].params[i]
	return param.name, param.typ
}

func (t *TypeSet) NumResults(ty Type) int {
	return len(t.nodes[ty].results)
}

func (t *TypeSet) Result(ty Type, i int) (name string, rty Type) {
	result := t.nodes[ty].results[i]
	return result.name, result.typ
}

func (t *TypeSet) getSlice(elt Type) Type {
	ty, node := t.newNode()
	node.kind = Slice
	node.elem = elt
	node.lsize = 1
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *TypeSet) getPointer(elt Type) Type {
	ty, node := t.newNode()
	node.kind = Pointer
	node.elem = elt
	node.lsize = 1
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *TypeSet) getChan(elt Type) Type {
	ty, node := t.newNode()
	node.kind = Chan
	node.elem = elt
	node.lsize = 1
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *TypeSet) getArray(elt Type, n int) Type {
	ty, node := t.newNode()
	node.kind = Array
	node.elem = elt
	node.lsize = t.nodes[elt].lsize*n + 1
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *TypeSet) getStruct(fields []named) Type {
	ty, node := t.newNode()
	node.kind = Struct
	node.fields = fields
	node.lsize = 1
	for _, f := range fields {
		node.lsize += t.nodes[f.typ].lsize
	}
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *TypeSet) getMap(kty, ety Type) Type {
	ty, node := t.newNode()
	node.kind = Map
	node.elem = ety
	node.key = kty
	node.lsize = 1 // like a pointer
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *TypeSet) getInterface(meths []named) Type {
	ty, node := t.newNode()
	node.kind = Interface
	sort.Slice(meths, func(i, j int) bool {
		return meths[i].name < meths[j].name
	})
	node.fields = meths
	node.lsize = 1 // like a pointer
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *TypeSet) getSignature(recv Type, params, results []named, variadic bool) Type {
	ty, node := t.newNode()
	node.kind = Func
	node.lsize = 1
	node.params = params
	node.results = results
	node.variadic = variadic
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *TypeSet) getTuple(elts []named) Type {
	ty, node := t.newNode()
	node.kind = Tuple
	node.lsize = 1
	node.fields = elts
	for _, field := range elts {
		node.lsize += t.nodes[field.typ].lsize
	}
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *TypeSet) getOrMake(ty Type, node *node) Type {
	ci := node.hash % uint32(cap(t.hash))
	ni := t.hash[ci]
	for ni != NoType {

		if t.equal(ni, ty) {
			t.nodes = t.nodes[:len(t.nodes)-1]
			return ni
		}
		ni = t.nodes[ni].next
	}
	node.next = t.hash[ty]
	t.hash[ci] = ty
	return ty
}

func (t *TypeSet) newNode() (Type, *node) {
	n := len(t.nodes)
	if n == cap(t.nodes) {
		t.grow()
	}
	t.nodes = t.nodes[:n+1]
	node := &t.nodes[n]
	node.zero()
	return Type(n), node
}

func (t *TypeSet) grow() {
	ncap := uint32(cap(t.nodes) * 2)
	tnodes := make([]node, len(t.nodes), ncap)
	thash := make([]Type, ncap)
	copy(tnodes, t.nodes)
	for i := range t.nodes {
		ty := Type(i)
		node := &tnodes[i]
		ci := node.hash % ncap
		node.next = Type(thash[ci])
		thash[ci] = ty
	}
	t.nodes = tnodes
	t.hash = thash
}

func (t *TypeSet) Equal(a, b Type) bool {
	return a == b
}

func (t *TypeSet) Len() int {
	return len(t.nodes)
}

func (t *TypeSet) equal(a, b Type) bool {
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
		return t.equal(anode.elem, bnode.elem)
	case Array:
		return t.equal(anode.elem, bnode.elem)
	case Struct, Interface, Tuple: // interface methods are sorted
		return t.namedsEqual(anode.fields, bnode.fields)
	case Map:
		return t.equal(anode.elem, bnode.elem) && t.equal(anode.key, bnode.key)
	case Func:
		if anode.variadic != bnode.variadic || anode.key != bnode.key {
			return false
		}
		return t.namedsEqual(anode.params, bnode.params) && t.namedsEqual(anode.results, bnode.results)
	default:
		panic("bad kind")
	}
	return true
}

func (t *TypeSet) namedsEqual(as, bs []named) bool {
	if len(as) != len(bs) {
		return false
	}
	for i := range as {
		anamed := as[i]
		bnamed := bs[i]
		if anamed.name != bnamed.name {
			return false
		}
		if anamed.loff != bnamed.loff {
			return false
		}
		if !t.equal(anamed.typ, bnamed.typ) {
			return false
		}
	}
	return true
}

func (t *TypeSet) hashCode(ty Type) uint32 {
	result := uint32(17)
	if ty < _endType {
		return uint32(_endType)
	}
	node := &t.nodes[ty]
	result ^= uint32(node.kind) << 3
	if node.elem != NoType {
		result ^= uint32(node.elem) << 5
	}
	if node.key != NoType {
		result ^= uint32(node.key) << 7
	}
	result ^= uint32(node.lsize) << 1
	if len(node.fields) > 0 {
		result = result ^ hashNamed(node.fields)
	}
	if len(node.params) > 0 {
		result = result ^ hashNamed(node.params)
	}
	if len(node.results) > 0 {
		result = result ^ hashNamed(node.results)
	}
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
