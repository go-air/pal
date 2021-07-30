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

package mem

import (
	"go/types"

	"github.com/go-air/pal/values"
)

type loc struct {
	class  Class
	root   T
	parent T

	vsz values.V

	in  []T
	out []T
}

type Model struct {
	locs   []loc
	values values.T
}

func NewModel(values values.T) *Model {
	res := &Model{
		// 0 -> not a mem
		// 1 -> zero mem
		locs:   make([]loc, 2, 128),
		values: values}
	zz := T(1)
	z := &res.locs[1]
	z.class = Zero
	z.parent = zz
	z.root = zz
	return res
}

func (mod *Model) Len() int {
	return len(mod.locs)
}

func (mod *Model) Access(m T, vs ...values.V) T {
	return T(0)
}

func (mod *Model) Zero() T {
	return T(1)
}

func (mod *Model) Local(ty types.Type) T {
	res := T(uint32(len(mod.locs)))
	mod.locs = append(mod.locs, loc{
		class:  Local,
		root:   res,
		parent: res,
		vsz:    mod.values.TypeSize(ty)})
	return res
}

func (mod *Model) Global(ty types.Type) T {
	res := T(uint32(len(mod.locs)))
	mod.locs = append(mod.locs, loc{
		class:  Global,
		root:   res,
		parent: res,
		vsz:    mod.values.TypeSize(ty)})
	return res
}

func (mod *Model) Heap(ty types.Type) T {
	res := T(uint32(len(mod.locs)))
	mod.locs = append(mod.locs, loc{
		class:  Heap,
		root:   res,
		parent: res,
		vsz:    mod.values.TypeSize(ty)})
	return res
}

func (mod *Model) Opaque(ty types.Type) T {
	res := T(uint32(len(mod.locs)))
	mod.locs = append(mod.locs, loc{
		class:  Opaque,
		root:   res,
		parent: res,
		vsz:    mod.values.TypeSize(ty)})
	return res
}
