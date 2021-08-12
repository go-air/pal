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

package ssa2pal

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"os"
	"sort"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/memory"
	"github.com/go-air/pal/results"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

type T struct {
	// pass object: analyzer is "github.com/go-air/pal".SSAAnalyzer()
	pass *analysis.Pass
	ssa  *buildssa.SSA
	// this is a state variable which
	// represents the current package under
	// analysis.
	pkg      *ssa.Package
	indexing indexing.T
	results  *results.T
	pkgres   *results.PkgRes
	buildr   *results.Builder
	// map from ssa.Value to memory locs

	vmap map[ssa.Value]memory.Loc

	funcs map[*ssa.Function]*Func
}

func New(pass *analysis.Pass, vs indexing.T) (*T, error) {
	palres := pass.Analyzer.FactTypes[0].(*results.T)
	pkgPath := pass.Pkg.Path()
	fmt.Printf("ssa2pal: %s\n", pkgPath)
	pkgRes := results.NewPkgRes(pkgPath, vs)
	for _, imp := range pass.Pkg.Imports() {
		iPath := imp.Path()
		//fmt.Printf("\t%s: importing %s\n", pkgPath, iPath)
		if palres.Lookup(iPath) == nil {
			return nil, fmt.Errorf("couldn't find pal results for %s\n", iPath)
		}
	}

	ssapkg := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	ssapkg.Pkg.Build()

	pal := &T{
		pass:     pass,
		pkg:      ssapkg.Pkg,
		ssa:      ssapkg,
		results:  palres,
		pkgres:   pkgRes,
		indexing: vs,
		buildr:   results.NewBuilder(pkgRes),
		vmap:     make(map[ssa.Value]memory.Loc, 8192),

		funcs: make(map[*ssa.Function]*Func)}
	return pal, nil
}

func (p *T) GenResult() (*results.T, error) {
	var err error
	mbrs := p.ssa.Pkg.Members
	mbrKeys := make([]string, 0, len(mbrs))
	// get and sort relevant member keys for determinism
	for name, mbr := range mbrs {
		switch mbr.Token() {
		case token.TYPE, token.CONST:
			continue
		}
		mbrKeys = append(mbrKeys, name)
	}
	sort.Strings(mbrKeys)
	// add globals
	for _, name := range mbrKeys {
		mbr := mbrs[name]
		switch g := mbr.(type) {
		case *ssa.Global:
			p.buildr.Reset()
			p.genGlobal(p.buildr, name, g)
		}
	}

	// add funcs
	for name, mbr := range p.ssa.Pkg.Members {
		switch fn := mbr.(type) {
		case *ssa.Function:
			p.buildr.Reset()
			if err = p.addFuncDecl(p.buildr, name, fn); err != nil {
				return nil, err
			}
		}
	}
	// TBD: calc results from generation above

	// place the results for current package in p.results.
	p.putResults()
	return p.results, nil
}

func (p *T) genGlobal(buildr *results.Builder, name string, x *ssa.Global) {
	// globals are in general pointers to the globally stored
	// index
	buildr.Pos = x.Pos()
	buildr.Class = memory.Global
	if token.IsExported(name) {
		// mark opaque because packages which import this one
		// may set the variable to whatever.
		buildr.Attrs = memory.IsOpaque
	}
	switch ty := x.Type().Underlying().(type) {
	case *types.Pointer:
		buildr.Type = ty.Elem()
		buildr.SrcKind = results.SrcVar

		loc, ptr := buildr.GenWithPointer()
		p.vmap[x] = ptr
		if traceLocVal {
			fmt.Printf("g %s %s %s %v %p\n", x.Name(), plain.String(loc), buildr.Type, x, x)
		}

		return

	default:
		msg := fmt.Sprintf(
			"unexpected ssa global member type for %s %T\n",
			name,
			buildr.Type)
		panic(msg)
	}
}

func (p *T) addFuncDecl(bld *results.Builder, name string, fn *ssa.Function) error {
	if fn.Signature.Variadic() {
		// XXX(wsc)
		return fmt.Errorf(
			"warning: \"%s\".%s: variadic params not yet handled\n",
			p.PkgPath(),
			name)
	}
	fmt.Printf("Func %s\n", name)
	memFn := NewFunc(bld, fn.Signature, name)
	p.vmap[fn] = memFn.Loc()
	bld.Reset()

	for i, param := range fn.Params {
		p.vmap[param] = memFn.ParamLoc(i)
	}
	// free vars not needed here -- top level func def

	p.funcs[fn] = memFn
	p.genBlocksValues(bld, name, fn)
	p.genConstraints(bld, name, fn)

	return nil
}

func (p *T) genBlocksValues(bld *results.Builder, name string, fn *ssa.Function) {
	for _, blk := range fn.Blocks {
		p.genBlockValues(bld, name, blk)
	}
	if fn.Recover != nil {
		p.genBlockValues(bld, name, fn.Recover)
	}
}

func (p *T) genConstraints(bld *results.Builder, name string, fn *ssa.Function) {
	for _, blk := range fn.Blocks {
		p.genBlockConstraints(bld, name, blk)
	}
	if fn.Recover != nil {
		p.genBlockConstraints(bld, name, fn.Recover)
	}
}

func (p *T) genBlockValues(bld *results.Builder, name string, blk *ssa.BasicBlock) {
	for _, i9n := range blk.Instrs {
		switch v := i9n.(type) {
		case *ssa.DebugRef, *ssa.Defer, *ssa.Go, *ssa.If, *ssa.Jump,
			*ssa.MapUpdate, *ssa.Panic, *ssa.Return,
			*ssa.RunDefers, *ssa.Send, *ssa.Store:
			// these are not values
		default:
			p.genValueLoc(bld, v.(ssa.Value))
		}
	}
}

func (p *T) genValueLoc(bld *results.Builder, v ssa.Value) memory.Loc {
	bld.Reset()
	bld.Pos = v.Pos()
	bld.Type = v.Type().Underlying()
	bld.Class = memory.Local
	var res memory.Loc
	switch v := v.(type) {
	case *ssa.Alloc:
		res = p.genAlloc(bld, v)
	case *ssa.Field:
		xloc, ok := p.vmap[v.X]
		if !ok {
			xloc = p.genValueLoc(bld, v.X)
		}
		res = bld.Field(xloc, v.Field)

	case *ssa.Index:
		xloc, ok := p.vmap[v.X]
		if !ok {
			xloc = p.genValueLoc(bld, v.X)
		}
		switch v := v.Index.(type) {
		case *ssa.Const:
			i64, ok := constant.Int64Val(v.Value)
			if !ok {
				panic("type checked const index not precise as int64")
			}
			i := int(i64) // should be ok also b/c it is type checked.
			res = bld.ArrayIndex(xloc, i)
		default:

			ty := v.Type().Underlying().(*types.Array)
			eltTy := ty.Elem()
			bld.Type = eltTy
			bld.Pos = v.Pos()
			res = bld.GenLoc()
			N := ty.Len()
			for i := int64(0); i < N; i++ {
				eltLoc := bld.ArrayIndex(xloc, int(i))
				bld.GenTransfer(res, eltLoc)
			}
		}

	default:
		res = bld.GenLoc()
	}
	p.vmap[v] = res
	return res
}

func (p *T) genAlloc(bld *results.Builder, a *ssa.Alloc) memory.Loc {
	bld.Reset()
	bld.Class = memory.Local
	if a.Heap {
		bld.Class = memory.Global
	}
	bld.Type = a.Type().Underlying().(*types.Pointer).Elem()
	bld.Pos = a.Pos()
	bld.SrcKind = results.SrcVar

	_, ptr := bld.GenWithPointer()
	return ptr
}

func (p *T) genBlockConstraints(bld *results.Builder, fnName string, blk *ssa.BasicBlock) error {

	for _, i9n := range blk.Instrs {
		if err := p.genI9nConstraints(bld, fnName, i9n); err != nil {
			return err
		}
	}
	return nil
}

func (p *T) genI9nConstraints(bld *results.Builder, fnName string, i9n ssa.Instruction) error {
	if traceGenI9n {
		fmt.Printf("gen %s\n", i9n)
	}
	switch i9n := i9n.(type) {
	case *ssa.Alloc: // done in gen locs
	case *ssa.BinOp:
	case *ssa.Call:
		p.call(bld, i9n.Call)
	case *ssa.ChangeInterface:
	case *ssa.ChangeType:
	case *ssa.Convert:
	case *ssa.DebugRef:
	case *ssa.Defer:
		p.call(bld, i9n.Call)
	case *ssa.Extract:
	case *ssa.Field: // done in gen locs

	case *ssa.FieldAddr:

		ptr := p.vmap[i9n.X]
		out := p.vmap[i9n]

		mdl := bld.Model()
		obj := mdl.Obj(ptr)
		fobj := memory.NoLoc
		if obj != memory.NoLoc {
			fobj = mdl.Field(obj, i9n.Field)
			bld.GenPointsTo(out, fobj)
			mdl.SetObj(out, fobj)
		} else {
			mdl.AddTransferIndex(out, ptr, i9n.Field)
		}

	case *ssa.Go:
		p.call(bld, i9n.Call)
	case *ssa.If:
	case *ssa.Index: // done in gen locs
	case *ssa.IndexAddr:
		ptr := p.vmap[i9n.X]
		p.buildr.Type = i9n.Type().Underlying()
		res := p.buildr.GenLoc()
		p.vmap[i9n] = res
		switch i9n.X.Type().Underlying().(type) {
		case *types.Pointer: // to array?
			p.buildr.GenTransfer(res, ptr)
		case *types.Slice:
			p.buildr.GenTransfer(res, ptr)
		default:
			panic("unexpected type of ssa.IndexAddr.X")
		}

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
		bld.Type = i9n.Type().Underlying().(*types.Slice).Elem()
		bld.Class = memory.Heap
		bld.SrcKind = results.SrcMakeSlice
		obj := bld.GenLoc()
		p.vmap[i9n] = obj
	case *ssa.MakeMap:
		ty := i9n.Type().Underlying().(*types.Map)
		keyType := ty.Key()
		valType := ty.Elem()
		bld.Reset()
		bld.Pos = i9n.Pos()
		bld.Type = keyType
		keyLoc := bld.GenLoc()
		bld.Type = valType
		valLoc := bld.GenLoc()
		_ = keyLoc
		_ = valLoc

	case *ssa.MapUpdate:

	case *ssa.Next: // either string iterator or map
		if !i9n.IsString {
			// not addressable

			return nil
		}
		// it is a map, type Tuple
	case *ssa.Panic:
	case *ssa.Phi:
		v := p.vmap[i9n]
		for _, x := range i9n.Edges {
			ev := p.vmap[x]
			bld.GenTransfer(v, ev)
		}
	case *ssa.Range:
	case *ssa.RunDefers:
		// no-op b/c we treat defers like calls.
	case *ssa.Select:
	case *ssa.Send:
	case *ssa.Return:
		var ssaFn *ssa.Function = i9n.Parent()
		var palFn *Func

		var ok bool
		palFn, ok = p.funcs[ssaFn]
		if !ok {
			return fmt.Errorf("couldn't find func %s\n", ssaFn.Name())
		}
		// copy results to palFn results...
		for i, res := range i9n.Results {
			resLoc := palFn.ResultLoc(i)
			// need to deal with things
			// which don't have pointers associated
			vloc := p.vmap[res]
			bld.GenTransfer(resLoc, vloc)
		}

	case *ssa.UnOp:
		// Load
		switch i9n.Op {
		case token.MUL:
			bld.GenLoad(p.vmap[i9n], p.vmap[i9n.X])
		default:
		}

	case *ssa.Slice:
	case *ssa.Store:
		vloc := p.vmap[i9n.Val]
		aloc := p.vmap[i9n.Addr]
		bld.GenStore(aloc, vloc)

	case *ssa.TypeAssert:
	default:
		panic("unknown ssa Instruction")
	}
	return nil
}

func (p *T) PkgPath() string {
	return p.pass.Pkg.Path()
}

func (p *T) call(b *results.Builder, c ssa.CallCommon) {
}

func (p *T) putResults() {
	if debugLogModel {
		fmt.Printf("built pal model for %s\n", p.pkgres.PkgPath)
		p.pkgres.PlainEncode(os.Stdout)
	}
	p.results.Put(p.pass.Pkg.Path(), p.pkgres)
}
