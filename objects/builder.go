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

type Builder struct {
	pkgPath  string
	indexing indexing.T
	mmod     *memory.Model
	ts       *typeset.TypeSet
	objects  []Object
	start    memory.Loc
	memGen   *memory.GenParams
}

func NewBuilder(pkgPath string, ind indexing.T, mem *memory.Model, ts *typeset.TypeSet) *Builder {
	b := &Builder{}
	b.pkgPath = pkgPath
	b.indexing = ind
	b.mmod = mem
	b.ts = ts
	b.memGen = memory.NewGenParams(ts)
	return b
}

func (b *Builder) Pos(pos token.Pos) *Builder {
	b.memGen.Pos(pos)
	return b
}

func (b *Builder) Class(c memory.Class) *Builder {
	b.memGen.Class(c)
	return b
}

func (b *Builder) Attrs(as memory.Attrs) *Builder {
	b.memGen.Attrs(as)
	return b
}

func (b *Builder) GoType(ty types.Type) *Builder {
	return b.Type(b.ts.FromGoType(ty))
}

func (b *Builder) Type(ty typeset.Type) *Builder {
	b.memGen.Type(ty)
	return b
}

func (b *Builder) Struct(gty *types.Struct) *Struct {
	b.GoType(gty)
	s := &Struct{}
	s.loc = b.mmod.Gen(b.memGen)
	s.typ = b.mmod.Type(s.loc)
	n := memory.Loc(b.mmod.LSize(s.loc))
	s.Fields = make([]memory.Loc, 0, gty.NumFields())
	for i := memory.Loc(1); i < n; i++ {
		pfloc := b.mmod.Parent(s.loc + i)
		if pfloc == s.loc {
			s.Fields = append(s.Fields, s.loc+i)
		}
	}
	if len(s.Fields) != gty.NumFields() {
		panic("internal error")
	}
	return s
}

func (b *Builder) Array(gty *types.Array) *Array {
	b.GoType(gty)
	a := &Array{}
	a.loc = b.mmod.Gen(b.memGen)
	a.typ = b.mmod.Type(a.loc)
	a.n = int(gty.Len()) // XXX int v int64
	a.elemSize = b.ts.Lsize(b.ts.Elem(a.typ))
	return a
}

func (b *Builder) Slice(gty *types.Slice, length, capacity indexing.I) *Slice {
	ty := b.ts.FromGoType(gty)
	b.Type(ty)
	s := &Slice{}
	s.Len = length
	s.Cap = capacity
	return s
}

func (b *Builder) AddSlot(slice *Slice, i indexing.I) {
	elem := b.ts.Elem(slice.typ)
	b.memGen.Type(elem)
	slice.Slots = append(slice.Slots, Slot{
		Loc: b.mmod.Gen(b.memGen),
		I:   i})
}

func (b *Builder) Map(gty *types.Map) *Map {
	ty := b.ts.FromGoType(gty)
	kty, ety := b.ts.Key(ty), b.ts.Elem(ty)
	mloc := b.mmod.Gen(b.memGen.Type(typeset.UnsafePointer))
	kloc := b.mmod.Gen(b.memGen.Type(kty))
	eloc := b.mmod.Gen(b.memGen.Type(ety))

	m := &Map{Key: kloc, Elem: eloc}
	m.loc = mloc
	m.typ = ty
	return m
}
