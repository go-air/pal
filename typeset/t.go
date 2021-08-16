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

import "sort"

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

func (t *T) Kind(ty Type) Kind {
	return t.nodes[ty].kind
}

func (t *T) ArrayLen(ty Type) int {
	node := &t.nodes[ty]
	n := node.lsize - 1
	eltSize := t.nodes[node.elem].lsize
	if n%eltSize != 0 {
		panic("bad array len")
	}
	return n / eltSize
}

func (t *T) NumFields(ty Type) int {
	return len(t.nodes[ty].fields)
}

func (t *T) Field(ty Type, i int) (name string, fty Type) {
	f := t.nodes[ty].fields[i]
	return f.name, f.typ
}

func (t *T) Recv(ty Type) Type {
	return t.nodes[ty].key
}

func (t *T) Variadic(ty Type) bool {
	return t.nodes[ty].variadic
}

func (t *T) NumParams(ty Type) int {
	return len(t.nodes[ty].params)
}

func (t *T) Param(ty Type, i int) (name string, pty Type) {
	param := t.nodes[ty].params[i]
	return param.name, param.typ
}

func (t *T) NumResults(ty Type) int {
	return len(t.nodes[ty].results)
}

func (t *T) Result(ty Type, i int) (name string, rty Type) {
	result := t.nodes[ty].results[i]
	return result.name, result.typ
}

func (t *T) getSlice(elt Type) Type {
	ty, node := t.newNode()
	node.kind = Slice
	node.elem = elt
	node.lsize = 1
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *T) getPointer(elt Type) Type {
	ty, node := t.newNode()
	node.kind = Pointer
	node.elem = elt
	node.lsize = 1
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *T) getChan(elt Type) Type {
	ty, node := t.newNode()
	node.kind = Chan
	node.elem = elt
	node.lsize = 1
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *T) getArray(elt Type, n int) Type {
	ty, node := t.newNode()
	node.kind = Chan
	node.elem = elt
	node.lsize = t.nodes[elt].lsize*n + 1
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *T) getStruct(fields []named) Type {
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

func (t *T) getMap(kty, ety Type) Type {
	ty, node := t.newNode()
	node.kind = Map
	node.elem = ety
	node.key = kty
	node.lsize = 1 // like a pointer
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *T) getInterface(meths []named) Type {
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

func (t *T) getSignature(recv Type, params, results []named, variadic bool) Type {
	ty, node := t.newNode()
	node.kind = Func
	node.lsize = 1
	node.params = params
	node.results = results
	node.variadic = variadic
	node.hash = t.hashCode(ty)
	return t.getOrMake(ty, node)
}

func (t *T) getTuple(elts []named) Type {
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

func (t *T) getOrMake(ty Type, node *node) Type {
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
	t.hash[ty] = ty
	return ty
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
	thash := make([]Type, ncap)
	copy(tnodes, t.nodes)
	for i := range t.nodes {
		ty := Type(i)
		node := &tnodes[i]
		ci := node.hash % ncap
		node.next = Type(thash[ci])
		thash[ci] = ty
	}
	t.nodes = t.nodes
	t.hash = thash
}

func (t *T) Equal(a, b Type) bool {
	return a == b
}

func (t *T) Len() int {
	return len(t.nodes)
}

func (t *T) equal(a, b Type) bool {
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
		if !t.equal(anamed.typ, bnamed.typ) {
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
