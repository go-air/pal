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
	"go/token"
	"go/types"

	"github.com/go-air/pal/typeset"
)

// GenParams encapsulates the information needed
// for a memory model to generate memory locations.
type GenParams struct {
	class Class
	attrs Attrs
	pos   token.Pos
	typ   typeset.Type

	ts *typeset.TypeSet
}

func NewGenParams(ts *typeset.TypeSet) *GenParams {
	gp := &GenParams{}
	gp.ts = ts
	return gp
}

func (p *GenParams) Reset() *GenParams {
	p.attrs = 0
	p.pos = token.NoPos
	p.typ = typeset.NoType
	p.class = 0
	return p
}

func (p *GenParams) Class(c Class) *GenParams {
	p.class = c
	return p
}

func (p *GenParams) Attrs(a Attrs) *GenParams {
	p.attrs = a
	return p
}

func (p *GenParams) Pos(pos token.Pos) *GenParams {
	p.pos = pos
	return p
}

func (p *GenParams) Type(t typeset.Type) *GenParams {
	p.typ = t
	return p
}

func (p *GenParams) GoType(t types.Type) *GenParams {
	p.typ = p.ts.FromGoType(t)
	return p
}
