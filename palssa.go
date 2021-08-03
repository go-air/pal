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
	"os"

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
	pkgres  *results.ForPkg
	buildr  *results.Builder
	imap    map[ssa.Instruction]memory.Loc
	vmap    map[ssa.Value]memory.Loc
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

	ssapkg := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	ssapkg.Pkg.Build()

	palSSA := &PalSSA{
		pass:    pass,
		pkg:     ssapkg.Pkg,
		ssa:     ssapkg,
		results: palres,
		pkgres:  pkgRes,
		values:  vs,
		buildr:  results.NewBuilder(pkgRes),
		imap:    make(map[ssa.Instruction]memory.Loc),
		vmap:    make(map[ssa.Value]memory.Loc)}
	return palSSA, nil
}

func (p *PalSSA) genResult() (*results.T, error) {
	var err error
	for name, mbr := range p.ssa.Pkg.Members {
		err = p.genMember(name, mbr)
		if err != nil {
			return nil, err
		}
	}
	p.putResults()
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
		p.genGlobal(buildr, name, x)
		return nil

	case *ssa.Function:
		p.addFuncDecl(buildr, name, x)
		return nil
	default:
		// NB(wsc) I think we can panic here...
		msg := fmt.Sprintf(
			"unexpected ssa member for %s %s %#v\n",
			name,
			x.Type(),
			x)
		return errors.New(msg)
	}
}

func (p *PalSSA) genGlobal(buildr *results.Builder, name string, x *ssa.Global) {
	// globals are in general pointers
	buildr.Pos = x.Pos()
	buildr.Type = x.Type()
	buildr.Class = memory.Global
	if token.IsExported(name) {
		buildr.Attrs = memory.IsOpaque
	}
	switch ty := buildr.Type.(type) {
	case *types.Pointer:
		buildr.Type = ty.Elem()
		buildr.SrcKind = results.SrcVar
		// gen what it points to
		dst := buildr.GenLoc()
		// pointer generated below
		buildr.Type = ty
		p := buildr.GenLoc()
		buildr.GenPointsTo(dst, p)

		return

	default:
		msg := fmt.Sprintf(
			"unexpected ssa global member type for %s %T\n",
			name,
			buildr.Type)
		_ = msg

		return // errors.New(msg)
	}
}

func (p *PalSSA) addFuncDecl(bld *results.Builder, name string, fn *ssa.Function) {
	if fn.Signature.Variadic() {
		// XXX(wsc)
		fmt.Printf(
			"warning: \"%s\".%s: variadic params not yet handled\n",
			p.PkgPath(),
			name)
		return
	}
	bld.Pos = fn.Pos()
	bld.Type = fn.Signature
	bld.Class = memory.Global
	bld.Attrs = memory.IsFunc
	bld.SrcKind = results.SrcFunc
	bld.GenLoc()
	bld.Reset()

	// handle parameters
	attrs := memory.IsParam
	if token.IsExported(name) {
		attrs |= memory.IsOpaque
	}
	for _, param := range fn.Params {
		bld.Pos = param.Pos()
		bld.Type = param.Type()
		bld.Attrs = attrs
		bld.Class = memory.Local
		bld.SrcKind = results.SrcVar
		bld.GenLoc()
	}
	// free vars not needed here -- top level func def

	// locals: *ssa.Alloc
	for _, a := range fn.Locals {
		bld.Reset()
		bld.Class = memory.Local
		if a.Heap {
			bld.Class = memory.Global
		}
		bld.Type = a.Type()
		bld.Pos = a.Pos()
		bld.SrcKind = results.SrcVar
		bld.GenLoc()
	}

	// blocks
	for _, blk := range fn.Blocks {
		p.genBlock(bld, name, blk)
	}
	if fn.Recover != nil {
		p.genBlock(bld, name, fn.Recover)
	}
}

func (p *PalSSA) genBlock(bld *results.Builder, fnName string, blk *ssa.BasicBlock) {
	for _, i9n := range blk.Instrs {
		p.genI9n(bld, fnName, i9n)
	}
}

func (p *PalSSA) genI9n(bld *results.Builder, fnName string, i9n ssa.Instruction) {
	rands := make([]ssa.Value, 0, 128)
	_ = rands
	bld.Pos = i9n.Pos()
	switch i9n := i9n.(type) {
	case *ssa.Alloc:
		bld.Type = i9n.Type()
		if i9n.Heap {
			bld.SrcKind = results.SrcNew
			bld.Class = memory.Heap
		} else {
			bld.SrcKind = results.SrcVar
			bld.Class = memory.Local
		}
		bld.GenLoc()
	case *ssa.BinOp:
	case *ssa.Call:
	case *ssa.ChangeInterface:
	case *ssa.ChangeType:
	case *ssa.Convert:
	case *ssa.DebugRef:
	case *ssa.Defer:
	case *ssa.Extract:
	case *ssa.Field:
	case *ssa.FieldAddr:
		// TBD: dependencies
		bld.Type = i9n.Type()
		bld.Class = memory.Local // really?
		bld.GenLoc()
	case *ssa.Go:
	case *ssa.If:
	case *ssa.Index:
	case *ssa.IndexAddr:
	case *ssa.Jump:
	case *ssa.Lookup:
	case *ssa.MakeInterface:
		bld.Type = i9n.Type()
		bld.Class = memory.Heap
		bld.SrcKind = results.SrcMakeInterface
		bld.GenLoc()
	case *ssa.MakeClosure:
	case *ssa.MakeChan:
	case *ssa.MakeSlice:
		bld.Type = i9n.Type()
		bld.Class = memory.Heap
		bld.SrcKind = results.SrcMakeSlice
		bld.GenLoc()
	case *ssa.MakeMap:
	case *ssa.MapUpdate:
	case *ssa.Next: // either string iterator or map
		if !i9n.IsString {
			// not addressable

			return
		}
		// map.
	case *ssa.Panic:
	case *ssa.Phi:
	case *ssa.Range:
	case *ssa.RunDefers:
	case *ssa.Select:
	case *ssa.Send:
	case *ssa.Return:
	case *ssa.UnOp:
	case *ssa.Slice:
	case *ssa.Store:
	case *ssa.TypeAssert:
	default:
		panic("unknown ssa Instruction")
	}
}

func (p *PalSSA) PkgPath() string {
	return p.pass.Pkg.Path()
}

func (p *PalSSA) putResults() {
	if debugLogModel {
		p.pkgres.MemModel.PlainEncode(os.Stdout)
	}
	p.results.Put(p.pass.Pkg.Path(), p.pkgres)
}
