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

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/xtruth"
)

// Type Model represents a memory model for a package.
type Model struct {
	locs        []loc
	constraints []Constraint
	indexing    indexing.T
}

// NewModel generates a new memory model for a package.
//
// index parameterises the resulting model on numeric
// (int) indexing.
func NewModel(index indexing.T) *Model {
	res := &Model{

		// 0 -> NoLoc
		// 1 -> Zero / nil/null
		locs:        make([]loc, 2, 1024),
		constraints: make([]Constraint, 0, 1024),
		indexing:    index}
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

func (mod *Model) Obj(ptr Loc) Loc {
	return mod.locs[ptr].obj
}

func (mod *Model) SetObj(ptr, dst Loc) {
	mod.locs[ptr].obj = dst
}

// Access returns the T which results from
// add vo to the virtual size of m.
func (mod *Model) Field(m Loc, i int) Loc {
	// @wsc: this can be done with a fuzzy binary
	// search
	n := m + 1 // first field
	for j := 0; j < i; j++ {
		sz := mod.locs[n].lsz
		isz, ok := mod.indexing.AsInt(sz)
		if !ok {
			return NoLoc
		}
		n += Loc(isz)
	}
	return n
}

func (mod *Model) ArrayIndex(m Loc, i int) Loc {
	n := m + 1
	sz := mod.locs[n].lsz
	isz, ok := mod.indexing.AsInt(sz)
	if !ok {
		return NoLoc

	}
	n += Loc(i * isz)
	return n
}

// VSize returns the virtual size of memory associated
// with m.
//
// The virtual size is the size according to the model,
// which is 1 + the sum of the the vsizes of all locations
// n such that mod.Parent(n) == m.
func (mod *Model) VSize(m Loc) indexing.I {
	return mod.locs[m].lsz
}

func (mod *Model) Overlaps(a, b Loc) xtruth.T {
	if mod.Root(a) != mod.Root(b) {
		return xtruth.False
	}
	if a == b {
		return xtruth.True
	}
	return xtruth.X
}

func (mod *Model) Equals(a, b Loc) xtruth.T {
	if a != b {
		return xtruth.False
	}
	return xtruth.X
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

// PlainCoderAt returns a plain.Coder for the information
// associated with memory at index i.
func (mod *Model) PlainCoderAt(i int) plain.Coder {
	return &mod.locs[i]
}

// Cap destructively changes the total size of mod.
//
func (mod *Model) Cap(c int) {
	if cap(mod.locs) < c {
		tmp := make([]loc, c)
		copy(tmp, mod.locs)
		mod.locs = tmp
	}
	mod.locs = mod.locs[:c]
}

// validates that
//
// n.vsz = 1 + Sum(c.vsz, c a child of n)
func (mod *Model) Check() error {
	N := len(mod.locs)
	sizes := make(map[Loc]int)
	for i := 2; i < N; i++ {
		loc := &mod.locs[i]
		if loc.parent == Loc(i) {
			continue
		}
		sz, ok := mod.indexing.AsInt(loc.lsz)
		if !ok {
			return fmt.Errorf("vsz not const")
		}
		ttl, _ := sizes[loc.parent]
		sizes[loc.parent] = sz + ttl
	}
	for m, sz := range sizes {
		sz++
		loc := &mod.locs[m]
		realsz, ok := mod.indexing.AsInt(loc.lsz)
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
	case *types.Signature:
		// a virtual place for the func
		mod.locs = append(mod.locs, l)
		*sum++

		// let's do this later...
		if ty.Variadic() {
			//panic("unimplemented variadic")
		}
		if ty.Recv() != nil {
			panic("unimplemented method")
		}

		parms := ty.Params()
		if parms != nil {
			N := parms.Len()
			for i := 0; i < N; i++ {
				p := parms.At(i)
				mod.add(p.Type(), class, attrs, n, r, sum)
			}
		}
		rets := ty.Results()
		if rets != nil {
			N := rets.Len()
			for i := 0; i < N; i++ {
				ret := rets.At(i)
				mod.add(ret.Type(), class, attrs, n, r, sum)
			}
		}
	case *types.Basic, *types.Pointer, *types.Interface:
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

	case *types.Tuple:
		mod.locs = append(mod.locs, l)
		*sum++
		tn := ty.Len()
		for i := 0; i < tn; i++ {
			mod.add(ty.At(i).Type(), class, attrs, n, r, sum)
		}
	case *types.Named:
		// no space reserved for named types, go to
		// underlying
		return mod.add(ty.Underlying(), class, attrs, p, r, sum)

	default:
		panic(fmt.Sprintf("%s: unexpected/unimplemented", ty))
	}
	// we added a slot at dst[n] for ty,  set it
	mod.locs[n].lsz = mod.indexing.FromInt(*sum - lastSum)
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
func (mod *Model) AddPointsTo(a, b Loc) {
	mod.constraints = append(mod.constraints, AddressOf(a, b))
}

func (mod *Model) GenWithPointer(ty types.Type, c Class, as Attrs) (obj, ptr Loc) {
	obj = mod.GenRoot(ty, c, as)
	ptr = Loc(len(mod.locs))

	mod.locs = append(mod.locs, loc{class: c, attrs: as, parent: ptr, root: ptr, obj: obj})
	mod.AddPointsTo(ptr, obj)
	return
}

// dst = *src
func (mod *Model) AddLoad(dst, src Loc) {
	mod.constraints = append(mod.constraints, Load(dst, src))
}

// *dst = src
func (mod *Model) AddStore(dst, src Loc) {
	mod.constraints = append(mod.constraints, Store(dst, src))
}

// dst = src
func (mod *Model) AddTransfer(dst, src Loc) {
	mod.constraints = append(mod.constraints, Transfer(dst, src))
}

func (mod *Model) AddTransferIndex(dst, src Loc, i indexing.I) {
	mod.constraints = append(mod.constraints, TransferIndex(dst, src, i))
}

func (mod *Model) Solve() {
	// apply constraints until fixed point.
}

// PointsTo places the points-to set of
// p in dst and returns it.
func (mod *Model) PointsToFor(dst []Loc, p Loc) []Loc {
	return dst
}

// Export exports the model 'mod', removing unnecessary local mem.Locs and
// compacting the result by permuting the remaining locations.  Export returns
// the permutation if 'perm' is non-nil.
//
// Generally, after Export is called, 'mod' contains no local variables.  One
// can retrieve points-to information for local variables using PoinstToFor,
// before calling Export.
func (mod *Model) Export(perm []Loc) []Loc {
	return perm
}

// Import imports 'other', merging it with
// mod in place.
func (mod *Model) Import(other *Model) {

}

func (mod *Model) PlainEncodeConstraints(w io.Writer) error {

	_, err := fmt.Fprintf(w, "%d\n", len(mod.constraints))
	if err != nil {
		return fmt.Errorf("mod:constraints: %w", err)

	}
	for i := range mod.constraints {
		c := &mod.constraints[i]
		if err := c.PlainEncode(w); err != nil {
			return err
		}
		_, err := fmt.Fprint(w, "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func (mod *Model) PlainDecodeConstraints(r io.Reader) error {
	var N int
	var err error
	_, err = fmt.Fscanf(r, "%d\n", &N)
	if err != nil {

		return fmt.Errorf("mod:constraints:N: %d %w", N, err)
	}
	mod.constraints = nil
	constraints := make([]Constraint, N)
	buf := make([]byte, 1)
	for i := 0; i < N; i++ {
		c := &constraints[i]
		err = c.PlainDecode(r)
		if err != nil {

			return fmt.Errorf("mod:constraints[%d]: %w", i, err)
		}
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return fmt.Errorf("mod:constraintsnl[%d]: %w", i, err)

		}
		if buf[0] != byte('\n') {
			return fmt.Errorf("unexpected '%s'", string(buf))
		}
	}
	mod.constraints = constraints
	return nil
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
	return mod.PlainEncodeConstraints(w)
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
	// OVFL?
	N := Loc(uint32(nLocsInt))
	if N > Loc(uint32(cap(mod.locs))) {
		tmp := make([]loc, N)
		mod.locs = tmp
	}
	mod.locs = mod.locs[:N]
	buf := make([]byte, 1)
	for i := Loc(0); i < N; i++ {
		p := &mod.locs[i]
		if err = p.PlainDecode(br); err != nil {
			return err
		}
		if _, err = io.ReadFull(br, buf); err != nil {
			return err
		}
		if buf[0] != '\n' {
			return fmt.Errorf("expected newline")
		}
	}
	return mod.PlainDecodeConstraints(r)
}
