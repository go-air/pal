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

package results

import (
	"go/token"
	"go/types"

	"github.com/go-air/pal/memory"
)

type Builder struct {
	Attrs   memory.Attrs
	Class   memory.Class
	Pos     token.Pos
	Type    types.Type
	SrcKind SrcKind

	pkg *PkgRes
	mdl *memory.Model
}

func NewBuilder(pkg *PkgRes) *Builder {
	return &Builder{pkg: pkg, mdl: pkg.MemModel}
}

func (b *Builder) Reset() {
	b.Attrs = memory.Attrs(0)
	b.Class = memory.Class(0)
	b.SrcKind = SrcKind(0)
	b.Pos = -1
	b.Type = nil
}

func (b *Builder) GenLoc() memory.Loc {
	res := b.mdl.GenRoot(b.Type, b.Class, b.Attrs)
	b.pkg.set(res, &SrcInfo{Kind: b.SrcKind, Pos: b.Pos})
	return res
}

func (b *Builder) Field(m memory.Loc, i int) memory.Loc {
	return b.mdl.Field(m, i)
}

func (b *Builder) ArrayIndex(m memory.Loc, i int) memory.Loc {
	return b.mdl.ArrayIndex(m, i)
}

func (b *Builder) GenWithPointer() (obj, ptr memory.Loc) {
	obj, ptr = b.mdl.GenWithPointer(b.Type, b.Class, b.Attrs)
	b.pkg.set(obj, &SrcInfo{Kind: b.SrcKind, Pos: b.Pos})
	b.pkg.set(ptr, &SrcInfo{Kind: b.SrcKind, Pos: b.Pos})
	return
}

func (b *Builder) GenPointsTo(dst, p memory.Loc) {
	b.mdl.AddPointsTo(dst, p)
}

func (b *Builder) GenStore(dst, src memory.Loc) {
	b.mdl.AddStore(dst, src)
}

func (b *Builder) GenLoad(dst, src memory.Loc) {
	b.mdl.AddLoad(dst, src)
}

func (b *Builder) GenTransfer(dst, src memory.Loc) {
	b.mdl.AddTransfer(dst, src)
}

func (b *Builder) Model() *memory.Model {
	return b.mdl

}

func (b *Builder) Check() error {
	return b.mdl.Check()
}
