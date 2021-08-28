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
	"go/token"
	"io"
	"strconv"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/typeset"
	"github.com/go-air/pal/xtruth"
)

// Type Model represents a memory model for a package.
type Model[Index plain.Coder] struct {
	locs        []loc
	constraints []Constraint[Index]
	indexing    indexing.T[Index]
	work        []Loc
}

// NewModel generates a new memory model for a package.
//
// index parameterises the resulting model on numeric
// (int) indexing.
func NewModel[Index plain.Coder](index indexing.T[Index]) *Model[Index] {
	res := &Model[Index]{

		// 0 -> NoLoc
		// 1 -> Zero / nil/null
		locs:        make([]loc, 2, 1024),
		constraints: make([]Constraint[Index], 0, 1024),
		indexing:    index}
	zz := Loc(1)
	z := &res.locs[1]
	z.class = Zero
	z.parent = zz
	z.root = zz
	return res
}

func (mod *Model[Index]) Len() int {
	return len(mod.locs)
}

func (mod *Model[Index]) At(i int) Loc {
	return Loc(i)
}

func (mod *Model[Index]) IsRoot(m Loc) bool {
	return mod.Parent(m) == m
}

func (mod *Model[Index]) Parent(m Loc) Loc {
	return mod.locs[m].parent
}

func (mod *Model[Index]) Root(m Loc) Loc {
	return mod.locs[m].root
}

func (mod *Model[Index]) Obj(ptr Loc) Loc {
	return mod.locs[ptr].obj
}

func (mod *Model[Index]) SetObj(ptr, dst Loc) {
	mod.locs[ptr].obj = dst
}

func (mod *Model[Index]) Pos(m Loc) token.Pos {
	return mod.locs[m].pos
}

// Access returns the T which results from
// add vo to the virtual size of m.
func (mod *Model[Index]) Field(m Loc, i int) Loc {
	// @wsc: this can be done with a fuzzy binary
	// search
	n := m + 1 // first field
	for j := 0; j < i; j++ {
		sz := mod.locs[n].lsz
		n += Loc(sz)
	}
	return n
}

// ArrayIndex returns the memory model location of `m` at index `i`.
func (mod *Model[Index]) ArrayIndex(m Loc, i int) Loc {
	n := m + 1
	sz := mod.locs[n].lsz
	n += Loc(i * sz)
	return n
}

// Lsize returns the virtual size of memory associated
// with m.
//
// The virtual size is the size according to the model,
// which is 1 + the sum of the the vsizes of all locations
// n such that mod.Parent(n) == m.
func (mod *Model[Index]) Lsize(m Loc) int {
	return mod.locs[m].lsz
}

func (mod *Model[Index]) Overlaps(a, b Loc) xtruth.T {
	if mod.Root(a) != mod.Root(b) {
		return xtruth.False
	}
	if a == b {
		return xtruth.True
	}
	return xtruth.X
}

func (mod *Model[Index]) Equals(a, b Loc) xtruth.T {
	if a != b {
		return xtruth.False
	}
	return xtruth.X
}

func (mod *Model[Index]) Zero() Loc {
	return Loc(1)
}

func (mod *Model[Index]) Type(m Loc) typeset.Type {
	return mod.locs[m].typ
}

func (mod *Model[Index]) Gen(gp *GenParams) Loc {
	var sum int
	p := Loc(uint32(len(mod.locs)))
	result := mod.add(gp, p, p, &sum)
	for _, ptr := range mod.work {
		gp.typ = gp.ts.Elem(mod.Type(ptr))
		sum := 0
		p = Loc(uint32(len(mod.locs)))
		obj := mod.add(gp, p, p, &sum)
		mod.locs[ptr].obj = obj
		mod.AddAddressOf(ptr, obj)
	}
	mod.work = mod.work[:0]
	return result
}

func (mod *Model[Index]) WithPointer(gp *GenParams) (obj, ptr Loc) {
	obj = mod.Gen(gp)
	ptr = Loc(uint32(len(mod.locs)))
	mod.locs = append(mod.locs, loc{
		class:  gp.class,
		attrs:  gp.attrs,
		pos:    gp.pos,
		typ:    gp.typ,
		parent: ptr,
		root:   ptr,
		obj:    obj})
	mod.AddAddressOf(ptr, obj)
	return
}

// PlainCoderAt returns a plain.Coder for the information
// associated with memory at index i.
func (mod *Model[Index]) PlainCoderAt(i int) plain.Coder {
	return &mod.locs[i]
}

// Cap destructively changes the total size of mod.
//
func (mod *Model[Index]) Cap(c int) {
	if cap(mod.locs) < c {
		tmp := make([]loc, c)
		copy(tmp, mod.locs)
		mod.locs = tmp
	}
	mod.locs = mod.locs[:c]
}

// add adds a root recursively according to ty.
//
// add is responsible for setting the size, parent, class, attrs, typ, and root
// of all added nodes.
//
func (mod *Model[Index]) add(gp *GenParams, p, r Loc, sum *int) Loc {
	n := Loc(uint32(len(mod.locs)))
	l := loc{
		parent: p,
		root:   r,
		class:  gp.class,
		attrs:  gp.attrs,
		pos:    gp.pos,
		typ:    gp.typ}
	lastSum := *sum
	added := true
	switch gp.ts.Kind(gp.typ) {
	// these are added as pointers here indirect associattions (params,
	// returns, ...) are done in github.com/go-air/objects.Builder
	case typeset.Basic, typeset.Pointer, typeset.Interface:
		mod.locs = append(mod.locs, l)
		*sum++
	case typeset.Slice, typeset.Chan, typeset.Map:
		mod.locs = append(mod.locs, l)
		*sum++

	case typeset.Array:
		mod.locs = append(mod.locs, l)
		*sum++

		m := gp.ts.ArrayLen(gp.typ)
		elemTy := gp.ts.Elem(gp.typ)
		for i := 0; i < m; i++ {
			gp.typ = elemTy
			mod.add(gp, n, r, sum)
		}
	case typeset.Struct:
		mod.locs = append(mod.locs, l)
		*sum++
		nf := gp.ts.NumFields(gp.typ)
		styp := gp.typ
		for i := 0; i < nf; i++ {
			_, gp.typ, _ = gp.ts.Field(styp, i)
			mod.add(gp, n, r, sum)
		}

	case typeset.Tuple:
		mod.locs = append(mod.locs, l)
		*sum++
		tn := gp.ts.NumFields(gp.typ)
		ttyp := gp.typ
		for i := 0; i < tn; i++ {
			_, gp.typ, _ = gp.ts.Field(ttyp, i)
			mod.add(gp, n, r, sum)
		}
	case typeset.Named:
		gp.typ = gp.ts.Underlying(gp.typ)
		mod.add(gp, p, r, sum)
		added = false
	case typeset.Func:
		mod.locs = append(mod.locs, l)
		*sum++
		fty := gp.typ
		rcvty := gp.ts.Recv(fty)
		if rcvty != typeset.NoType {
			gp.typ = gp.ts.PointerTo(rcvty)
			mod.work = append(mod.work, mod.add(gp, n, r, sum))
		}
		// TBD FreeVars

		np := gp.ts.NumParams(fty)
		for i := 0; i < np; i++ {
			_, gp.typ = gp.ts.Param(fty, i)
			gp.typ = gp.ts.PointerTo(gp.typ)
			mod.work = append(mod.work, mod.add(gp, n, r, sum))
		}
		nr := gp.ts.NumResults(fty)
		for i := 0; i < nr; i++ {
			_, gp.typ = gp.ts.Result(fty, i)
			gp.typ = gp.ts.PointerTo(gp.typ)
			mod.work = append(mod.work, mod.add(gp, n, r, sum))
		}

	default:
		panic(fmt.Sprintf("%d: unexpected/unimplemented", gp.typ))
	}
	if added {
		// we added a slot at dst[n] for ty,  set its size
		mod.locs[n].lsz = *sum - lastSum
		return n
	}
	return NoLoc
}

func (mod *Model[Index]) Attrs(m Loc) Attrs {
	return mod.locs[m].attrs
}

func (mod *Model[Index]) AddAttrs(m Loc, a Attrs) {
	mm := &mod.locs[m]
	mm.attrs |= a
}

func (mod *Model[Index]) SetAttrs(m Loc, a Attrs) {
	mm := &mod.locs[m]
	mm.attrs = a
}

// a = &b
func (mod *Model[Index]) AddAddressOf(a, b Loc) {
	mod.constraints = append(mod.constraints, AddressOf[Index](a, b))
}

// dst = *src
func (mod *Model[Index]) AddLoad(dst, src Loc) {
	mod.constraints = append(mod.constraints, Load[Index](dst, src))
}

// *dst = src
func (mod *Model[Index]) AddStore(dst, src Loc) {
	mod.constraints = append(mod.constraints, Store[Index](dst, src))
}

// dst = src
func (mod *Model[Index]) AddTransfer(dst, src Loc) {
	mod.AddTransferIndex(dst, src, mod.indexing.Zero())
}

func (mod *Model[Index]) AddTransferIndex(dst, src Loc, i Index) {
	mod.constraints = append(mod.constraints, TransferIndex(dst, src, i))
}

func (mod *Model[Index]) Solve() {
	// apply constraints until fixed point.
}

// PointsTo places the points-to set of
// p in dst and returns it.
func (mod *Model[Index]) PointsToFor(dst []Loc, p Loc) []Loc {
	return dst
}

// Export exports the model 'mod', removing unnecessary local mem.Locs and
// compacting the result by permuting the remaining locations.  Export returns
// the permutation if 'perm' is non-nil.
//
// Generally, after Export is called, 'mod' contains no local variables.  One
// can retrieve points-to information for local variables using PoinstToFor,
// before calling Export.
func (mod *Model[Index]) Export(perm []Loc) []Loc {
	return perm
}

// Import imports 'other', merging it with
// mod in place.
func (mod *Model[Index]) Import(other *Model[Index]) {

}

func (mod *Model[Index]) PlainEncodeConstraints(w io.Writer) error {

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

func (mod *Model[Index]) PlainDecodeConstraints(r io.Reader) error {
	var N int
	var err error
	_, err = fmt.Fscanf(r, "%d\n", &N)
	if err != nil {

		return fmt.Errorf("mod:constraints:N: %d %w", N, err)
	}
	mod.constraints = nil
	constraints := make([]Constraint[Index], N)
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

func (mod *Model[Index]) PlainEncode(w io.Writer) error {
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

func (mod *Model[Index]) PlainDecode(r io.Reader) error {
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
