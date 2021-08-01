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

	"github.com/go-air/pal/mem"
)

type Builder struct {
	Attrs   mem.Attrs
	Class   mem.Class
	Pos     token.Pos
	ty      types.Type
	SrcKind SrcKind

	pkg *Pkg
}

func NewBuilder(pkg *Pkg) *Builder {
	return &Builder{pkg: pkg}
}

func (b *Builder) Reset() {
	b.Attrs = mem.Attrs(0)
	b.Class = mem.Class(0)
	b.SrcKind = SrcKind(0)
	b.Pos = -1
	b.ty = nil
}

func (b *Builder) GenLoc() mem.Loc {
	res := b.pkg.MemModel.Add(b.ty, b.Class, b.Attrs)
	b.pkg.set(res, &SrcInfo{Kind: b.SrcKind, Pos: b.Pos})
	return res
}
