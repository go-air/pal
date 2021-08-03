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

package memory

import (
	"bufio"
	"fmt"
	"go/types"
	"io"
	"strconv"

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/values"
)

// Type Model represents a memory model for a package.
type Model struct {
	locs   []loc
	values values.T
}

// NewModel generates a new memory model for a package.
//
// values parameterises the resulting model on numeric
// (int) values.
func NewModel(values values.T) *Model {
	res := &Model{
		// 0 -> not a mem
		// 1 -> zero mem
		locs:   make([]loc, 2, 1024),
		values: values}
	zz := Loc(1)
	z := &res.locs[1]
	z.class = Zero
	z.parent = zz
	z.root = zz
	return res
}

func (mod *Model) Len() int {
	return len(mod.locs)
}

func (mod *Model) At(i int) Loc {
	return Loc(i)
}

func (mod *Model) IsRoot(m Loc) bool {
	return mod.Parent(m) == m
}

func (mod *Model) Parent(m Loc) Loc {
	return mod.locs[m].parent
}

func (mod *Model) Root(m Loc) Loc {
	return mod.locs[m].root
}

// Access returns the T which results from
// add vo to the virtual size of m.
func (mod *Model) Access(m Loc, vo values.V) Loc {
	return Loc(0)
}

// ?(wsco) pass values.V instead?
func (mod *Model) AccessField(m Loc, i int) (Loc, error) {
	msz, ok := mod.values.AsInt(mod.VSize(m))
	if !ok {
		panic("")
	}
	_ = msz
	//fmt.Printf("%d.access field %d, len %d vsz %d end %d\n", m, i, mod.Len(), msz, Loc(msz)+m)
	n := m + 1 // first field

	for j := 0; j < i; j++ {
		sz := mod.locs[n].vsz
		isz, ok := mod.values.AsInt(sz)
		if !ok {
			return Loc(0), fmt.Errorf("memory model has non-const field size: %s", plain.String(m))
		}
		//fmt.Printf("\tfield %d sz %d\n", j, isz)
		n += Loc(isz)
	}
	return n, nil
}

func (mod *Model) VSize(m Loc) values.V {
	return mod.locs[m].vsz
}

func (mod *Model) Overlaps(a, b Loc) values.AbsTruth {
	if mod.Root(a) != mod.Root(b) {
		return values.False
	}
	if a == b {
		return values.True
	}
	return values.Unknown
}

func (mod *Model) Equals(a, b Loc) values.AbsTruth {
	if a != b {
		return values.False
	}
	return values.Unknown
}

func (mod *Model) Zero() Loc {
	return Loc(1)
}

func (mod *Model) Local(ty types.Type, attrs Attrs) Loc {
	var sum int
	p := Loc(uint32(len(mod.locs)))
	return mod.add(ty, Local, attrs, p, p, &sum)
}

func (mod *Model) Global(ty types.Type, attrs Attrs) Loc {
	var sum int
	p := Loc(uint32(len(mod.locs)))
	return mod.add(ty, Global, attrs, p, p, &sum)
}

func (mod *Model) Heap(ty types.Type, attrs Attrs) Loc {
	var sum int
	p := Loc(uint32(len(mod.locs)))
	return mod.add(ty, Heap, attrs, p, p, &sum)
}

// GenRoot generates a new root memory location.
func (mod *Model) GenRoot(ty types.Type, class Class, attrs Attrs) Loc {
	var sum int
	p := Loc(uint32(len(mod.locs)))
	return mod.add(ty, class, attrs, p, p, &sum)
}

func (mod *Model) PlainCoderAt(i int) plain.Coder {
	return &mod.locs[i]
}

func (mod *Model) Cap(c int) {
	if cap(mod.locs) < c {
		tmp := make([]loc, c)
		copy(tmp, mod.locs)
		mod.locs = tmp
	}
	mod.locs = mod.locs[:c]
}

func (mod *Model) Check() error {
	N := len(mod.locs)
	sizes := make(map[Loc]int)
	for i := 2; i < N; i++ {
		loc := &mod.locs[i]
		if loc.parent == Loc(i) {
			continue
		}
		sz, ok := mod.values.AsInt(loc.vsz)
		if !ok {
			return fmt.Errorf("vsz not const")
		}
		ttl, _ := sizes[loc.parent]
		sizes[loc.parent] = sz + ttl
	}
	for m, sz := range sizes {
		sz++
		loc := &mod.locs[m]
		realsz, ok := mod.values.AsInt(loc.vsz)
		if !ok {
			return fmt.Errorf("vsz not const")
		}
		if sz != realsz {
			return fmt.Errorf("loc %s sum %d size %d\n", plain.String(m), sz, realsz)
		}
	}
	return nil
}

// add adds a root recursively according to ty.
//
// add is responsible for setting the size, parent, class, attrs, and root
// of all added nodes.
//
func (mod *Model) add(ty types.Type, class Class, attrs Attrs, p, r Loc, sum *int) Loc {
	n := Loc(uint32(len(mod.locs)))
	l := loc{
		parent: p,
		root:   r,
		class:  class,
		attrs:  attrs}
	lastSum := *sum
	switch ty := ty.(type) {
	case *types.Basic, *types.Pointer, *types.Signature, *types.Interface:
		mod.locs = append(mod.locs, l)
		*sum++
	case *types.Array:
		mod.locs = append(mod.locs, l)
		*sum++
		m := int(ty.Len())
		for i := 0; i < m; i++ {
			mod.add(ty.Elem(), class, attrs, n, r, sum)
		}
	case *types.Map:
		mod.locs = append(mod.locs, l)
		*sum++
		mod.add(ty.Key(), class, attrs, n, r, sum)
		mod.add(ty.Elem(), class, attrs, n, r, sum)
	case *types.Struct:
		mod.locs = append(mod.locs, l)
		*sum++
		nf := ty.NumFields()
		for i := 0; i < nf; i++ {
			fty := ty.Field(i).Type()
			mod.add(fty, class, attrs, n, r, sum)
		}
	case *types.Slice:
		mod.locs = append(mod.locs, l)
		*sum++
		mod.add(ty.Elem(), class, attrs, n, r, sum)
	case *types.Chan:
		mod.locs = append(mod.locs, l)
		*sum++
		mod.add(ty.Elem(), class, attrs, n, r, sum)

	case *types.Named:
		// no space reserved for named types, go to
		// underlying
		return mod.add(ty.Underlying(), class, attrs, p, r, sum)

	default:
		panic(fmt.Sprintf("%s: unexpected/unimplemented", ty))
	}
	// we added a slot at dst[n] for ty,  set it
	mod.locs[n].vsz = mod.values.FromInt(*sum - lastSum)
	return n
}

func (mod *Model) Attrs(m Loc) Attrs {
	return mod.locs[m].attrs
}

func (mod *Model) AddAttrs(m Loc, a Attrs) {
	mm := &mod.locs[m]
	mm.attrs |= a
}

func (mod *Model) SetAttrs(m Loc, a Attrs) {
	mm := &mod.locs[m]
	mm.attrs = a
}

// a = &b
func (mod *Model) GenPointsTo(a, b Loc) {
	loc := &mod.locs[a]
	loc.pointsTo = append(loc.pointsTo, b)
}

// dst = *src
func (mod *Model) GenLoad(dst, src Loc) {
	loc := &mod.locs[dst]
	loc.loads = append(loc.loads, src)

}

// *dst = src
func (mod *Model) GenStore(dst, src Loc) {
	loc := &mod.locs[dst]
	loc.stores = append(loc.stores, src)

}

// dst = src
func (mod *Model) GenTransfer(dst, src Loc) {
	loc := &mod.locs[dst]
	_ = loc
}

func (mod *Model) Solve() {
	// apply constraints until fixed point.
}

// PointsTo places the points-to set of
// p in dst and returns it.
func (mod *Model) PointsToFor(dst []Loc, p Loc) []Loc {
	return dst
}

// Export exports the model 'mod', removing
// unnecessary local mem.Locs and compacting
// the result by permuting the remaining
// locations.  Export returns the permutation
// if 'perm' is non-nil.
//
// Generally, after Export is called,
// 'mod' contains no local variables.
// One can retrieve points-to information
// for local variables using PoinstToFor,
// before calling Export.
func (mod *Model) Export(perm []Loc) []Loc {
	return perm
}

// Import imports 'other', merging it with
// mod in place.
func (mod *Model) Import(other *Model) {

}

func (mod *Model) PlainEncode(w io.Writer) error {
	fmt.Fprintf(w, "%d locs\n", len(mod.locs))
	var err error
	for i := range mod.locs {
		m := &mod.locs[i]
		err = m.PlainEncode(w)
		if err != nil {
			return err
		}
		_, err := w.Write([]byte{byte('\n')})
		if err != nil {
			return err
		}
	}
	return nil
}

func (mod *Model) PlainDecode(r io.Reader) error {
	br := bufio.NewReader(r)
	nLocs, err := br.ReadString('\n')
	if err != nil {
		return err
	}
	nLocsInt, err := strconv.ParseUint(nLocs, 10, 32)
	if err != nil {
		return err
	}
	// OVFL
	N := Loc(uint32(nLocsInt))
	if N > Loc(uint32(cap(mod.locs))) {
		tmp := make([]loc, N)
		mod.locs = tmp
	}
	mod.locs = mod.locs[:N]
	for i := Loc(0); i < N; i++ {
		p := &mod.locs[i]
		if err = p.PlainDecode(br); err != nil {
			return err
		}
	}
	return nil
}
