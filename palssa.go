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
	"fmt"
	"go/token"
	"go/types"

	"github.com/go-air/pal/memory"
	"github.com/go-air/pal/results"
	"github.com/go-air/pal/values"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

type PalSSA struct {
	pass *analysis.Pass
	ssa  *buildssa.SSA
	// this is a state variable which
	// represents the current package under
	// analysis.
	pkg     *ssa.Package
	values  values.T
	results *results.T
	pkgres  *results.Pkg
	buildr  *results.Builder
}

func NewPalSSA(pass *analysis.Pass, vs values.T) (*PalSSA, error) {
	palres := pass.Analyzer.FactTypes[0].(*results.T)
	pkgPath := pass.Pkg.Path()
	pkgRes := results.NewPkg(pkgPath, vs)
	for _, imp := range pass.Pkg.Imports() {
		iPath := imp.Path()
		//fmt.Printf("\t%s: importing %s\n", pkgPath, iPath)
		if palres.Lookup(iPath) == nil {
			return nil, fmt.Errorf("couldn't find pal results for %s", iPath)
		}
	}

	ssa := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	ssa.Pkg.Build()

	palSSA := &PalSSA{
		pass:    pass,
		pkg:     ssa.Pkg,
		ssa:     ssa,
		results: palres,
		pkgres:  pkgRes,
		values:  vs,
		buildr:  results.NewBuilder(pkgRes)}
	return palSSA, nil
}

func (p *PalSSA) genResult() (*results.T, error) {
	for name, mbr := range p.ssa.Pkg.Members {
		p.genMember(name, mbr)
	}
	p.putResults()
	return p.results, nil
}

func (p *PalSSA) genMember(name string, mbr ssa.Member) {
	switch mbr.Token() {
	case token.TYPE, token.CONST:
		return
	}
	buildr := p.buildr
	buildr.Reset()
	fmt.Printf("genMember for \"%s\".%s\n", p.pkg.Pkg.Path(), name)

	switch x := mbr.(type) {
	case *ssa.Global:
		buildr.Pos = x.Pos()
		buildr.Type = x.Type()
		buildr.Class = memory.Global
		switch buildr.Type.(type) {
		case *types.Basic, *types.Array, *types.Struct, *types.Map, *types.Chan, *types.Pointer:
			buildr.SrcKind = results.SrcVar
		case *types.Signature:
			fmt.Printf("gotta func global %s\n", name)
			buildr.SrcKind = results.SrcFunc
			buildr.Attrs = memory.IsFunc
		default:
			fmt.Printf("gotta unknown %s\n", name, buildr.Type)
		}

	case *ssa.Function:
		fmt.Printf("gotta func member %s %s %#v\n", name, x.Type(), x)
		buildr.Pos = x.Pos()
		buildr.Type = x.Signature
		buildr.Class = memory.Global
		buildr.Attrs = memory.IsFunc
	}
	buildr.GenLoc()
}

func (p *PalSSA) putResults() {
	p.results.Put(p.pass.Pkg.Path(), p.pkgres)
}

/*
func (f *FromSSA) registerAlloc(a *ssa.Alloc) mem.T {
	if !a.Heap {
		return mem.T(0)
	}
	var m mem.T
	m = f.mems.Heap(a.Type(), 0)
	f.set(m, &i)
	return m
}

func (f *FromSSA) registerGlobal(g *ssa.Global) mem.T {
	var m mem.T
	var i = MemSSA{Pkg: f.pkg, Global: g}
	m = f.mems.Global(g.Type(), 0)
	f.set(m, &i)
	return m
}
*/