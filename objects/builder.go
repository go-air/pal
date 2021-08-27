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
	omap     map[memory.Loc]Object // map memory locs to canonical objects
	start    memory.Loc            // after import, what is our minimum?
	mgp      *memory.GenParams     // memory loc generation parameters
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
	m := b.GoType(gty).Gen()
	b.walkObj(m)
	return b.omap[m].(*Struct)
}

func (b *Builder) Tuple(ty *types.Tuple) *Tuple {
	m := b.GoType(ty).Gen()
	b.walkObj(m)
	return b.omap[m].(*Tuple)
}

func (b *Builder) Array(gty *types.Array) *Array {
	m := b.GoType(gty).Gen()
	b.walkObj(m)
	return b.omap[m].(*Array)
}

func (b *Builder) Slice(gty *types.Slice, length, capacity indexing.I) *Slice {
	ty := b.ts.FromGoType(gty)
	m := b.Type(ty).Gen()
	b.walkObj(m)
	sl := b.omap[m].(*Slice)
	sl.Len = length
	sl.Cap = capacity
	b.AddSlot(sl, b.indexing.Zero())
	return sl
}

func (b *Builder) AddSlot(slice *Slice, i indexing.I) {
	elemTy := b.ts.Elem(slice.typ)
	ptrTy := b.ts.PointerTo(elemTy)
	ptr := b.Type(ptrTy).Gen()
	obj := b.Type(elemTy).Gen()
	b.walkObj(obj)
	b.AddTransferIndex(ptr, slice.loc, i)
	b.AddAddressOf(ptr, obj)

	slice.slots = append(slice.slots, Slot{
		Ptr: ptr,
		Obj: obj,
		I:   i})
}

func (b *Builder) Map(gty *types.Map) *Map {
	ty := b.ts.FromGoType(gty)
	m := b.Type(ty).Gen()
	b.walkObj(m)
	return b.omap[m].(*Map)
}

func (b *Builder) Chan(gty *types.Chan) *Chan {
	ty := b.ts.FromGoType(gty)
	m := b.Type(ty).Gen()
	b.walkObj(m)
	return b.omap[m].(*Chan)
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
	_, fn.loc = b.Type(fn.typ).WithPointer()

	b.mmod.AddAddressOf(fn.loc, fn.loc)

	fn.params = make([]memory.Loc, sig.Params().Len())
	fn.variadic = sig.Variadic()
	fn.results = make([]memory.Loc, sig.Results().Len())

	b.Class(memory.Local) // for all params and returns

	recv := sig.Recv()
	var obj memory.Loc

	if recv != nil {
		b.mgp.Type(b.ts.FromGoType(recv.Type()))
		obj, fn.recv = b.mmod.WithPointer(b.mgp)
		b.walkObj(obj)
	}
	params := sig.Params()
	N := params.Len()
	for i := 0; i < N; i++ {
		param := params.At(i)
		pty := b.ts.FromGoType(param.Type().Underlying())
		obj, fn.params[i] =
			b.Pos(param.Pos()).Type(pty).Attrs(memory.IsParam | opaque).WithPointer()
		b.walkObj(obj)
	}
	rets := sig.Results()
	N = rets.Len()
	for i := 0; i < N; i++ {
		ret := rets.At(i)
		rty := b.ts.FromGoType(ret.Type())
		obj, fn.results[i] =
			b.Pos(ret.Pos()).Type(rty).Attrs(memory.IsReturn | opaque).WithPointer()
		b.walkObj(obj)
	}
	// TBD: FreeVars
	b.omap[fn.loc] = fn
	return fn
}

// second pass to associate objects with all object like memory locations...
// the input is likely to just create roots at variables, but we need objects
// everywhere...
func (b *Builder) walkObj(m memory.Loc) {
	ty := b.mmod.Type(m)
	//fmt.Printf("walkObj %s ty %s\n", plain.String(m), b.ts.String(ty))
	ki := b.ts.Kind(ty)
	switch ki {
	case typeset.Basic:
		switch ty {
		case typeset.Pointer:
			if b.omap[m] == nil {
				ptr := &Pointer{}
				ptr.loc = m
				ptr.typ = ty
				b.omap[m] = ptr
			}
		}
	case typeset.Array:
		var arr *Array
		obj := b.omap[m]
		if obj == nil {
			arr = &Array{}
			arr.loc = m
			arr.typ = ty
			arr.n = int64(b.ts.ArrayLen(ty))
			arr.elemSize = int64(b.ts.Lsize(b.ts.Elem(ty)))
			b.omap[m] = arr
		} else {
			arr = obj.(*Array)
		}
		n := b.ts.ArrayLen(ty)
		for i := 0; i < n; i++ {
			elt := arr.At(i)
			b.walkObj(elt)
		}

	case typeset.Struct:
		var strukt *Struct
		obj := b.omap[m]
		if obj == nil {
			strukt = &Struct{}
			strukt.loc = m
			strukt.typ = ty
			b.omap[m] = strukt
		} else {
			strukt = obj.(*Struct)
		}
		n := b.ts.NumFields(ty)
		strukt.fields = make([]memory.Loc, n)
		for i := 0; i < n; i++ {
			_, _, foff := b.ts.Field(ty, i)
			floc := m + memory.Loc(foff)
			strukt.fields[i] = floc
			b.walkObj(floc)
		}

	case typeset.Chan:
		var ch *Chan
		obj := b.omap[m]
		if obj == nil {
			ch = &Chan{}
			ch.loc = m
			ch.typ = ty
			ch.slot = b.Type(b.ts.Elem(ty)).Gen()
			b.mmod.AddAddressOf(ch.loc, ch.slot)
			b.omap[m] = ch
		} else {
			ch = obj.(*Chan)
		}
		b.walkObj(ch.slot)
	case typeset.Map:
		var ma *Map
		obj := b.omap[m]
		if obj == nil {
			ma = &Map{}
			ma.loc = m
			ma.typ = ty
			ma.key = b.Type(b.ts.Key(ty)).Gen()
			ma.elem = b.Type(b.ts.Elem(ty)).Gen()
			b.mmod.AddAddressOf(ma.loc, ma.key)
			b.mmod.AddAddressOf(ma.loc, ma.elem)
			b.omap[m] = ma
		} else {
			ma = obj.(*Map)
		}
		b.walkObj(ma.key)
		b.walkObj(ma.elem)
	case typeset.Slice:
		var slice *Slice
		obj := b.omap[m]
		if obj == nil {
			slice = &Slice{}
			slice.loc = m
			slice.typ = b.mmod.Type(m)
			slice.Len = b.indexing.Zero()
			slice.Cap = b.indexing.Zero()
			b.AddAddressOf(slice.loc, b.mmod.Zero())
			b.omap[m] = slice
		}

	case typeset.Interface:
	case typeset.Func:
	case typeset.Named:
		panic("named foo")

	case typeset.Tuple:
		var tuple *Tuple
		obj := b.omap[m]
		if obj == nil {
			tuple = &Tuple{}
			tuple.loc = m
			tuple.typ = ty
			b.omap[m] = tuple
		} else {
			tuple = obj.(*Tuple)
		}
		n := b.ts.NumFields(ty)
		tuple.fields = make([]memory.Loc, n)
		for i := 0; i < n; i++ {
			_, _, foff := b.ts.Field(ty, i)
			floc := m + memory.Loc(foff)
			tuple.fields[i] = floc
			b.walkObj(floc)
		}
	}
}
