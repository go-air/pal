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

package mem

import (
	"fmt"
	"go/types"

	"github.com/go-air/pal/values"
)

type loc struct {
	class  Class
	attrs  Attrs
	root   T
	parent T

	vsz values.V

	// constraints
	pointsTo  []T // this loc points to that
	transfers []T //
	loads     []T // this loc = *(that loc)
	stores    []T // *(this loc) = that loc

	// points-to (and from)
	in  []T
	out []T
}

type Model struct {
	locs   []loc
	values values.T
	vszs   []int
}

func NewModel(values values.T) *Model {
	res := &Model{
		// 0 -> not a mem
		// 1 -> zero mem
		locs:   make([]loc, 2, 128),
		values: values}
	zz := T(1)
	z := &res.locs[1]
	z.class = Zero
	z.parent = zz
	z.root = zz
	return res
}

func (mod *Model) Len() int {
	return len(mod.locs)
}

func (mod *Model) At(i int) T {
	return T(i)
}

func (mod *Model) IsRoot(m T) bool {
	return mod.Parent(m) == m
}

func (mod *Model) Parent(m T) T {
	return mod.locs[m].parent
}

// Access returns the T which results from
// add vo to the virtual size of m.
func (mod *Model) Access(m T, vo values.V) T {
	return T(0)
}

func (mod *Model) VSize(m T) values.V {
	return mod.locs[m].vsz
}

func (mod *Model) Equals(a, b T) values.AbsTruth {
	if a != b {
		return values.False
	}
	return values.Unknown
}

func (mod *Model) Zero() T {
	return T(1)
}

func (mod *Model) Local(ty types.Type, attrs Attrs) T {
	var sum int
	p := T(uint32(len(mod.locs)))
	return mod.add(ty, Local, attrs, p, p, &sum)
}

func (mod *Model) Global(ty types.Type, attrs Attrs) T {
	var sum int
	p := T(uint32(len(mod.locs)))
	return mod.add(ty, Global, attrs, p, p, &sum)
}

func (mod *Model) Heap(ty types.Type, attrs Attrs) T {
	var sum int
	p := T(uint32(len(mod.locs)))
	return mod.add(ty, Heap, attrs, p, p, &sum)
}

func (mod *Model) add(ty types.Type, class Class, attrs Attrs, p, r T, sum *int) T {
	n := T(uint32(len(mod.locs)))
	l := loc{
		parent: p,
		root:   r,
		class:  class,
		attrs:  attrs}
	lastSum := *sum
	switch ty := ty.(type) {
	case *types.Basic, *types.Pointer:
		mod.locs = append(mod.locs, l)
		*sum++
	case *types.Array:
		mod.locs = append(mod.locs, l)
		m := int(ty.Len())
		for i := 0; i < m; i++ {
			mod.add(ty.Elem(), class, attrs, n, r, sum)
		}
	case *types.Map:
		mod.locs = append(mod.locs, l)
		mod.add(ty.Key(), class, attrs, n, r, sum)
		mod.add(ty.Elem(), class, attrs, n, r, sum)
	case *types.Struct:
		mod.locs = append(mod.locs, l)
		nf := ty.NumFields()
		for i := 0; i < nf; i++ {
			fty := ty.Field(i).Type()
			mod.add(fty, class, attrs, n, r, sum)
		}
	case *types.Named:
		// no space reserved for named types, go to
		// underlying
		return mod.add(ty.Underlying(), class, attrs, p, r, sum)

	default:
		panic(fmt.Sprintf("%s: unexpected/unimplemented", ty))
	}
	// we added a slot at dst[n] for ty, set it
	mod.locs[n].vsz = mod.values.FromInt(*sum - lastSum)
	return n
}

func (mod *Model) Attrs(m T) Attrs {
	return mod.locs[m].attrs
}

func (mod *Model) AddAttrs(m T, a Attrs) {
	mm := &mod.locs[m]
	mm.attrs |= a
}

func (mod *Model) SetAttrs(m T, a Attrs) {
	mm := &mod.locs[m]
	mm.attrs = a
}

// a = &b
func (mod *Model) GenPointsTo(a, b T) {
	loc := &mod.locs[a]
	loc.pointsTo = append(loc.pointsTo, b)
}

// dst = *src
func (mod *Model) GenLoad(dst, src T) {
	loc := &mod.locs[dst]
	loc.loads = append(loc.loads, src)

}

// *dst = src
func (mod *Model) GenStore(dst, src T) {
	loc := &mod.locs[dst]
	loc.stores = append(loc.stores, src)

}

// dst = src
func (mod *Model) GenTransfer(dst, src T) {
	loc := &mod.locs[dst]
	loc.stores = append(loc.stores, src)
}

func (mod *Model) Solve() {
	// apply constraints until fixed point.
}

// PointsTo places the points-to set of
// p in dst and returns it.
func (mod *Model) PointsToFor(dst []T, p T) []T {
	return dst
}

// Export exports the model 'mod', removing
// unnecessary local mem.Ts and compacting
// the result by permuting the remaining
// locations.  Export returns the permutation
// if 'perm' is non-nil.
//
// Generally, after Export is called,
// 'mod' contains no local variables.
// One can retrieve points-to information
// for local variables using PoinstToFor,
// before calling Export.
func (mod *Model) Export(perm []T) []T {
	return perm
}

// Import imports 'other', merging it with
// mod in place.
func (mod *Model) Import(other *Model) {

}
