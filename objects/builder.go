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

	class   memory.Class
	attrs   memory.Attrs
	pos     token.Pos
	srcKind memory.SrcKind
	typ     typeset.Type
}

func NewBuilder(pkgPath string, ind indexing.T, mem *memory.Model, ts *typeset.TypeSet) *Builder {
	b := &Builder{}
	b.pkgPath = pkgPath
	b.indexing = ind
	b.mmod = mem
	b.ts = ts
	return b
}

func (b *Builder) Reset() *Builder {
	b.class = memory.Class(0)
	b.attrs = memory.Attrs(0)
	b.pos = token.NoPos
	b.srcKind = memory.SrcKind(0)
	b.typ = typeset.NoType
	return b
}

func (b *Builder) Class(c memory.Class) *Builder {
	b.class = c
	return b
}

func (b *Builder) Attrs(a memory.Attrs) *Builder {
	b.attrs = a
	return b
}

func (b *Builder) Pos(p token.Pos) *Builder {
	b.pos = p
	return b
}

func (b *Builder) Kind(k memory.SrcKind) *Builder {
	b.srcKind = k
	return b
}

func (b *Builder) Type(t typeset.Type) *Builder {
	b.typ = t
	return b
}

func (b *Builder) GoType(t types.Type) *Builder {
	b.typ = b.ts.FromGoType(t)
	return b
}
