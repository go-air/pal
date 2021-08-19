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

package objects

import (
	"go/token"
	"go/types"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/memory"
	"github.com/go-air/pal/typeset"
)

// Builder is a type which supports coordinating
// higher level Go objects (funcs, slices, maps, ...)
// with the lower level memory model and typeset.
type Builder struct {
	pkgPath  string
	indexing indexing.T
	mmod     *memory.Model
	ts       *typeset.TypeSet
	omap     map[memory.Loc]Object
	start    memory.Loc
	mgp      *memory.GenParams
}

// NewBuilder creates a new builder from a package Path
// and an indexing.T which manages representing indexing
// expressiosn.
func NewBuilder(pkgPath string, ind indexing.T) *Builder {
	b := &Builder{}
	b.pkgPath = pkgPath
	b.indexing = ind
	b.mmod = memory.NewModel(ind)
	b.ts = typeset.New()
	b.mgp = memory.NewGenParams(b.ts)
	b.omap = make(map[memory.Loc]Object)
	return b
}

func (b *Builder) Memory() *memory.Model {
	return b.mmod
}

func (b *Builder) TypeSet() *typeset.TypeSet {
	return b.ts
}

func (b *Builder) AddAddressOf(ptr, obj memory.Loc) {
	b.mmod.AddAddressOf(ptr, obj)
}

func (b *Builder) AddLoad(dst, src memory.Loc) {
	b.mmod.AddLoad(dst, src)
}

func (b *Builder) AddStore(dst, src memory.Loc) {
	b.mmod.AddStore(dst, src)
}

func (b *Builder) AddTransfer(dst, src memory.Loc) {
	b.mmod.AddTransfer(dst, src)
}

func (b *Builder) AddTransferIndex(dst, src memory.Loc, i indexing.I) {
	b.mmod.AddTransferIndex(dst, src, i)
}

func (b *Builder) Pos(pos token.Pos) *Builder {
	b.mgp.Pos(pos)
	return b
}

func (b *Builder) Class(c memory.Class) *Builder {
	b.mgp.Class(c)
	return b
}

func (b *Builder) Attrs(as memory.Attrs) *Builder {
	b.mgp.Attrs(as)
	return b
}

func (b *Builder) GoType(ty types.Type) *Builder {
	return b.Type(b.ts.FromGoType(ty))
}

func (b *Builder) Type(ty typeset.Type) *Builder {
	b.mgp.Type(ty)
	return b
}

func (b *Builder) Struct(gty *types.Struct) *Struct {
	b.GoType(gty)
	s := &Struct{}
	s.loc = b.Gen()
	s.typ = b.mmod.Type(s.loc)
	n := memory.Loc(b.mmod.Lsize(s.loc))
	s.fields = make([]memory.Loc, 0, gty.NumFields())
	for i := memory.Loc(1); i < n; i++ {
		pfloc := b.mmod.Parent(s.loc + i)
		if pfloc == s.loc {
			s.fields = append(s.fields, s.loc+i)
		}
	}
	if len(s.fields) != gty.NumFields() {
		panic("internal error")
	}
	b.omap[s.loc] = s
	return s
}

func (b *Builder) Array(gty *types.Array) *Array {
	b.GoType(gty)
	a := &Array{}
	a.loc = b.Gen()
	a.typ = b.mmod.Type(a.loc)
	a.n = gty.Len()
	a.elemSize = int64(b.ts.Lsize(b.ts.Elem(a.typ)))
	b.omap[a.loc] = a
	return a
}

func (b *Builder) Slice(gty *types.Slice, length, capacity indexing.I) *Slice {
	ty := b.ts.FromGoType(gty)
	b.Type(ty)
	s := &Slice{}
	s.loc = b.Gen()
	s.Len = length
	s.Cap = capacity
	// add one uni-slot by default.
	b.AddSlot(s, b.indexing.Var())
	b.omap[s.loc] = s
	return s
}

func (b *Builder) AddSlot(slice *Slice, i indexing.I) {
	elem := b.ts.Elem(slice.typ)
	slice.slots = append(slice.slots, Slot{
		Loc: b.Type(elem).Gen(),
		I:   i})
}

func (b *Builder) Map(gty *types.Map) *Map {
	ty := b.ts.FromGoType(gty)
	kty, ety := b.ts.Key(ty), b.ts.Elem(ty)
	mloc := b.mmod.Gen(b.mgp.Type(ty))
	kloc := b.Type(kty).Gen()
	eloc := b.Type(ety).Gen()
	b.mmod.AddAddressOf(mloc, eloc)

	m := &Map{key: kloc, elem: eloc}
	m.loc = mloc
	m.typ = ty
	b.omap[m.loc] = m
	return m
}

func (b *Builder) Object(m memory.Loc) Object {
	return b.omap[m]
}

func (b *Builder) Gen() memory.Loc {
	return b.mmod.Gen(b.mgp)
}

func (b *Builder) WithPointer() (obj, ptr memory.Loc) {
	obj, ptr = b.mmod.WithPointer(b.mgp)
	return
}

func (b *Builder) Pointer(gtype *types.Pointer) *Pointer {
	ptr := &Pointer{}
	ptr.typ = b.ts.FromGoType(gtype)
	ptr.loc = b.Type(ptr.typ).Gen()
	b.omap[ptr.loc] = ptr
	return ptr
}

// Func makes a function object.  It is for top level functions
// which may or may not be declared.  `declName` must be empty
// iff the associated function is not declared.
func (b *Builder) Func(sig *types.Signature, declName string, opaque memory.Attrs) *Func {
	fn := &Func{declName: declName}
	fn.typ = b.ts.FromGoType(sig)
	fn.loc = b.Type(fn.typ).Gen()

	fn.params = make([]memory.Loc, sig.Params().Len())
	fn.variadic = sig.Variadic()
	fn.results = make([]memory.Loc, sig.Results().Len())

	b.Class(memory.Local) // for all params and returns

	recv := sig.Recv()

	if recv != nil {
		b.mgp.Type(b.ts.FromGoType(recv.Type()))
		fn.recv = b.mmod.Gen(b.mgp)
	}
	params := sig.Params()
	N := params.Len()
	for i := 0; i < N; i++ {
		param := params.At(i)
		pty := b.ts.FromGoType(param.Type())
		fn.params[i] =
			b.Pos(param.Pos()).Type(pty).Attrs(memory.IsParam | opaque).Gen()
	}
	rets := sig.Results()
	N = rets.Len()
	for i := 0; i < N; i++ {
		ret := rets.At(i)
		rty := b.ts.FromGoType(ret.Type())
		fn.results[i] =
			b.Pos(ret.Pos()).Type(rty).Attrs(memory.IsReturn | opaque).Gen()
	}
	// TBD: FreeVars
	b.omap[fn.loc] = fn
	return fn

}
