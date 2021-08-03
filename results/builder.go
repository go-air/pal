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

	pkg *ForPkg
}

func NewBuilder(pkg *ForPkg) *Builder {
	return &Builder{pkg: pkg}
}

func (b *Builder) Reset() {
	b.Attrs = memory.Attrs(0)
	b.Class = memory.Class(0)
	b.SrcKind = SrcKind(0)
	b.Pos = -1
	b.Type = nil
}

func (b *Builder) GenLoc() memory.Loc {
	res := b.pkg.MemModel.Gen(b.Type, b.Class, b.Attrs)
	b.pkg.set(res, &SrcInfo{Kind: b.SrcKind, Pos: b.Pos})
	return res
}

func (b *Builder) GenPointsTo(dst, p memory.Loc) {
	b.pkg.MemModel.GenPointsTo(dst, p)
}

func (b *Builder) GenStore(dst, src memory.Loc) {
	b.pkg.MemModel.GenStore(dst, src)
}

func (b *Builder) Check() error {
	return b.pkg.MemModel.Check()
}
