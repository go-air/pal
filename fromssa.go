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

// This file provides support for generating and using the core pal
// functions from golang.org/s/tools/go/ssa representation.

package pal

import (
	"math"

	"github.com/go-air/pal/values"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

type MemSSA struct {
	Pkg *ssa.Package

	// these fields are like a union
	// only at most one is non-nil
	Param         *ssa.Parameter
	Global        *ssa.Global
	Alloc         *ssa.Alloc
	MakeChan      *ssa.MakeChan
	MakeClosure   *ssa.MakeClosure
	MakeInterface *ssa.MakeInterface
	MakeMap       *ssa.MakeMap
	MakeSlice     *ssa.MakeSlice
}

type FromSSA struct {
	ssa *buildssa.SSA
	// this is a state variable which
	// represents the current package under
	// analysis.
	pkg  *ssa.Package
	mems *Mems
	info []MemSSA // indexed by Mem
}

func NewFromSSA(b *buildssa.SSA, vs values.T) *FromSSA {
	mems := NewMems(vs)
	return &FromSSA{ssa: b, mems: mems, info: make([]MemSSA, mems.Len(), 128)}
}

func (f *FromSSA) startPackage(p *ssa.Package) {
	f.pkg = p
}

func (f *FromSSA) endPackage(p *ssa.Package) {
	f.pkg = nil
}

func (f *FromSSA) Info(m Mem) *MemSSA {
	return &f.info[m]
}

func (f *FromSSA) registerParam(p *ssa.Parameter) Mem {
	m := f.mems.Opaque(p.Type())
	f.set(m, &MemSSA{Pkg: f.pkg, Param: p})
	return m
}

func (f *FromSSA) registerAlloc(a *ssa.Alloc) Mem {
	if !a.Heap {
		return Mem(0)
	}
	var m Mem
	var i = MemSSA{Pkg: f.pkg, Alloc: a}
	m = f.mems.Heap(a.Type())
	f.set(m, &i)
	return m
}

func (f *FromSSA) registerGlobal(g *ssa.Global) Mem {
	var m Mem
	var i = MemSSA{Pkg: f.pkg, Global: g}
	m = f.mems.Global(g.Type())
	f.set(m, &i)
	return m
}

func (f *FromSSA) set(m Mem, info *MemSSA) {
	n := Mem(uint32(cap(f.info)))
	if m < n {
		f.info[m] = *info
		return
	}
	if m > math.MaxUint32/2 {
		n = math.MaxUint32
	} else {
		for n <= m {
			n *= 2
		}
	}
	infos := make([]MemSSA, n, n)
	copy(infos, f.info)
	infos[m] = *info
	f.info = infos
}
