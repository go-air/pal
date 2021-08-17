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

	start memory.Loc

	memGen *memory.GenParams
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
	b.memGen.Type(b.ts.FromGoType(ty))
	return b
}

func (b *Builder) Struct(gty *types.Struct) *Struct {
	b.GoType(gty)
	s := &Struct{}
	return s
}

func (b *Builder) Array(gty *types.Array) *Array {
	a := &Array{}
	return a
}

func (b *Builder) Slice(gty *types.Slice) *Slice {
	s := &Slice{}
	return s
}

func (b *Builder) Map(gty *types.Map) *Map {
	m := &Map{}
	return m
}
