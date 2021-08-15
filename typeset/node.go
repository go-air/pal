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

type node struct {
	kind    Kind
	elem    Type    // pointer, array, slice
	key     Type    // map only
	lsize   int     // memory model logical size
	fields  []named // struct only
	params  []named // name == nil ok
	results []named // name == nil ok

	// hashing
	next Type
	hash uint32
}

// for fields, parameters, named results
type named struct {
	name string
	typ  Type
}

func (n *node) zero() {
	n.kind = Basic
	n.elem = NoType
	n.key = NoType
	n.lsize = 1
	n.fields = nil
	n.params = nil
	n.results = nil
	n.hash = 0
	n.next = NoType
}

var basicNodes = []node{
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // NoType
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Bool
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Uint8
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Uint16
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Uint32
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Uint64
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Int8
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Int16
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Int32
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Int64
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Float32
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Float64
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Complex64
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Complex128
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // String
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // UnsafePointer
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}} // Uintptr
