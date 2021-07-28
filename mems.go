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

package pal

import (
	"go/types"

	"github.com/go-air/pal/values"
)

type mem struct {
	class  MemClass
	root   Mem
	parent Mem

	ty  types.Type
	vsz values.V

	in  []Mem
	out []Mem
}

type Mems struct {
	mems   []mem
	values values.T
}

func NewMems(values values.T) *Mems {
	return &Mems{
		// 0 -> not a mem
		// 1 -> zero mem
		mems:   make([]mem, 2, 128),
		values: values}
}

func (ms *Mems) Access(m Mem, vs ...values.V) Mem {
	return Mem(0)
}

func (ms *Mems) Zero() Mem {
	return Mem(1)
}

func (ms *Mems) Local(ty types.Type) Mem {
	res := Mem(uint32(len(ms.mems)))
	ms.mems = append(ms.mems, mem{
		class:  Local,
		root:   res,
		parent: res,
		ty:     ty,
		vsz:    ms.values.FromType(ty)})
	return res
}

func (ms *Mems) Global(ty types.Type) Mem {
	res := Mem(uint32(len(ms.mems)))
	ms.mems = append(ms.mems, mem{
		class:  Global,
		root:   res,
		parent: res,
		ty:     ty,
		vsz:    ms.values.FromType(ty)})
	return res
}

func (ms *Mems) Heap(ty types.Type) Mem {
	res := Mem(uint32(len(ms.mems)))
	ms.mems = append(ms.mems, mem{
		class:  Heap,
		root:   res,
		parent: res,
		ty:     ty,
		vsz:    ms.values.FromType(ty)})
	return res
}

func (ms *Mems) Opaque(ty types.Type) Mem {
	res := Mem(uint32(len(ms.mems)))
	ms.mems = append(ms.mems, mem{
		class:  Opaque,
		root:   res,
		parent: res,
		ty:     ty,
		vsz:    ms.values.FromType(ty)})
	return res
}
