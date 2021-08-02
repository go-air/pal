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
	"errors"
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
	defer p.putResults()
	var err error
	for name, mbr := range p.ssa.Pkg.Members {
		err = p.genMember(name, mbr)
		if err != nil {
			return nil, err
		}
	}
	return p.results, nil
}

func (p *PalSSA) genMember(name string, mbr ssa.Member) error {
	switch mbr.Token() {
	case token.TYPE, token.CONST:
		return nil
	}
	buildr := p.buildr
	buildr.Reset()

	switch x := mbr.(type) {
	case *ssa.Global:
		buildr.Pos = x.Pos()
		buildr.Type = x.Type()
		buildr.Class = memory.Global
		switch ty := buildr.Type.(type) {
		case *types.Pointer:
			buildr.Type = ty.Elem()
			buildr.SrcKind = results.SrcVar

		case *types.Basic, *types.Array, *types.Struct, *types.Map, *types.Chan:
			fmt.Printf("gotta basic %s\n", buildr.Type)
			buildr.SrcKind = results.SrcVar
		case *types.Signature:
			fmt.Printf("gotta func global %s\n", name)
			buildr.SrcKind = results.SrcFunc
			buildr.Attrs = memory.IsFunc
		default:
			msg := fmt.Sprintf(
				"unexpected ssa global member type for %s %T\n",
				name,
				buildr.Type)
			return errors.New(msg)
		}

	case *ssa.Function:
		fmt.Printf("gotta func member %s %s %#v\n", name, x.Type(), x)
		buildr.Pos = x.Pos()
		buildr.Type = x.Signature
		buildr.Class = memory.Global
		buildr.Attrs = memory.IsFunc
	default:
		msg := fmt.Sprintf(
			"unexpected ssa member for %s %s %#v\n",
			name,
			x.Type(),
			x)
		return errors.New(msg)
	}
	buildr.GenLoc()
	return nil
}

func (p *PalSSA) putResults() {
	p.results.Put(p.pass.Pkg.Path(), p.pkgres)
}
