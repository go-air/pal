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
	"errors"
	"fmt"
	"go/constant"
	"go/token"
	"go/types"

	"github.com/go-air/pal/index"
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
	pkg     *ssa.Package
	index   index.T
	results *results.T
	pkgres  *results.PkgRes
	buildr  *results.Builder
	// map from ssa.Value to memory locs
	vmap map[ssa.Value]memory.Loc
	// map from ssa.Value of type pointer
	// to the memory location of the unique object to which
	// vmap[p] points (clearly not all pointers point to
	// unique ops, but this is used for example to implement
	// *ssa.FieldAddr(*ssa.Alloc, n).
	omap map[ssa.Value]memory.Loc
	//
	funcs map[*ssa.Function]*Func
}

func New(pass *analysis.Pass, vs index.T) (*T, error) {
	palres := pass.Analyzer.FactTypes[0].(*results.T)
	pkgPath := pass.Pkg.Path()
	if pkgPath == "internal/cpu" {
		return nil, fmt.Errorf("skip %s", pkgPath)
	}
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
		pass:    pass,
		pkg:     ssapkg.Pkg,
		ssa:     ssapkg,
		results: palres,
		pkgres:  pkgRes,
		index:   vs,
		buildr:  results.NewBuilder(pkgRes),
		vmap:    make(map[ssa.Value]memory.Loc, 8192),
		omap:    make(map[ssa.Value]memory.Loc, 8192),
		funcs:   make(map[*ssa.Function]*Func)}
	return pal, nil
}

func (p *T) GenResult() (*results.T, error) {
	var err error
	// for every member generate memory locations
	// and constraints.
	for name, mbr := range p.ssa.Pkg.Members {
		err = p.genMember(name, mbr)
		if err != nil {
			return nil, err
		}
	}
	// TBD: calc results from generation above

	// place the results for current package in p.results.
	p.putResults()
	return p.results, nil
}

func (p *T) genMember(name string, mbr ssa.Member) error {
	switch mbr.Token() {
	case token.TYPE, token.CONST:
		return nil
	}
	buildr := p.buildr
	buildr.Reset()

	switch x := mbr.(type) {
	case *ssa.Global:
		// globals are treated
		p.genGlobal(buildr, name, x)
		return nil

	case *ssa.Function:
		return p.addFuncDecl(buildr, name, x)
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

func (p *T) genGlobal(buildr *results.Builder, name string, x *ssa.Global) {
	// globals are in general pointers to the globally stored
	// index
	buildr.Pos = x.Pos()
	buildr.Type = x.Type().Underlying()
	buildr.Class = memory.Global
	if token.IsExported(name) {
		// mark opaque because packages which import this one
		// may set the variable to whatever.
		buildr.Attrs = memory.IsOpaque
	}
	switch ty := buildr.Type.(type) {
	case *types.Pointer:
		buildr.Type = ty.Elem()
		buildr.SrcKind = results.SrcVar
		// gen what it points to
		// pointer generated below
		buildr.Type = ty

		loc, ptr := buildr.GenWithPointer()
		p.vmap[x] = ptr
		p.omap[x] = loc
		if traceLocVal {
			fmt.Printf("g %s %s %s\n", x.Name(), plain.String(loc), buildr.Type)
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

	// need to do this for result below
	p.funcs[fn] = memFn

	// locals: *ssa.Alloc
	for _, a := range fn.Locals {
		if _, present := p.vmap[a]; present {
			panic("double local?")
			continue
		}
		bld.Reset()
		bld.Class = memory.Local
		if a.Heap {
			bld.Class = memory.Global
		}
		bld.Type = a.Type().Underlying().(*types.Pointer).Elem()
		bld.Pos = a.Pos()
		bld.SrcKind = results.SrcVar
		obj, ptr := bld.GenWithPointer()
		p.vmap[a] = ptr
		p.omap[a] = obj

		if traceLocVal {
			fmt.Printf("l %s %s %s\n", a.Name(), plain.String(ptr), a.Type())
		}
	}

	// blocks
	for _, blk := range fn.Blocks {
		if err := p.genBlock(bld, name, blk); err != nil {
			return err
		}
	}
	if fn.Recover != nil {
		if err := p.genBlock(bld, name, fn.Recover); err != nil {
			return err
		}
	}
	return nil
}

func (p *T) genBlock(bld *results.Builder, fnName string, blk *ssa.BasicBlock) error {
	for _, i9n := range blk.Instrs {
		if err := p.genI9n(bld, fnName, i9n); err != nil {
			return err
		}
	}
	return nil
}

func (p *T) genI9n(bld *results.Builder, fnName string, i9n ssa.Instruction) error {
	if traceGenI9n {
		fmt.Printf("gen %s\n", i9n)
	}
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("on i9n %s in %s\n", i9n, fnName)
			panic(e)
		}
	}()
	bld.Pos = i9n.Pos()
	switch i9n := i9n.(type) {
	case *ssa.Alloc:
		if _, present := p.vmap[i9n]; present {
			// we batch created the locals...
			return nil
		}
		bld.Pos = i9n.Pos()
		bld.Type = i9n.Type().Underlying().(*types.Pointer).Elem().Underlying()
		if i9n.Heap {
			bld.SrcKind = results.SrcNew
			bld.Class = memory.Heap
		} else {
			bld.SrcKind = results.SrcVar
			bld.Class = memory.Local
		}
		obj, ptr := bld.GenWithPointer()
		p.vmap[i9n] = ptr
		p.omap[i9n] = obj
		if traceLocVal {
			fmt.Printf("a %s %s\n", i9n.Name(), plain.String(ptr))
		}
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
	case *ssa.Field:
		xloc := p.getLoc(bld, i9n.X)
		floc := bld.Field(xloc, i9n.Field)
		p.vmap[i9n] = floc

	case *ssa.FieldAddr:
		// the requirements are subtle.  we need to be able
		// to calculate deref(i9n.X) to get
		elemTy := i9n.X.Type().Underlying().(*types.Pointer).Elem()
		_ = elemTy

		var dobj memory.Loc
		fmt.Printf("FieldAdd: X=%v\n", i9n.X)
		var ok bool
		if dobj, ok = p.omap[i9n.X]; !ok {
			if true {
				// we need to make sure other ops
				// which can lead to this are modelled

				panic(fmt.Sprintf("&o.f o=%s i9n %s\n", i9n.X, i9n))
			}
			bld.Reset()
			elemTy := i9n.X.Type().Underlying().(*types.Pointer).Elem()
			bld.Type = elemTy
			bld.Class = memory.Local
			bld.Pos = i9n.Pos()
			dobj = bld.GenLoc()
			bld.GenLoad(dobj, p.getLoc(bld, i9n.X))
		}
		//ptr := bld.GenPointerTo(dobj)
		//p.vmap[i9n] = ptr
		p.omap[i9n] = dobj
	case *ssa.Go:
		p.call(bld, i9n.Call)
	case *ssa.If:
	case *ssa.Index:
		// i9n.X is array.
		// if i9n.Index is constant, we can
		// access its model
		//
		// if not, perhaps we back off with transfer
		// constraints and a new Loc
		xloc := p.getLoc(bld, i9n.X)
		switch idx := i9n.Index.(type) {
		case *ssa.Const:
			i64, ok := constant.Int64Val(idx.Value)
			if !ok {
				panic("type checked const index not precise as int64")
			}
			i := int(i64) // should be ok also b/c it is type checked.
			eltLoc := bld.ArrayIndex(xloc, i)
			p.vmap[i9n] = eltLoc
		default:

			ty := i9n.Type().Underlying().(*types.Array)
			N := ty.Len()
			bld.Type = ty
			bld.Pos = i9n.Pos()
			res := bld.GenLoc()
			for i := int64(0); i < N; i++ {
				eltLoc := bld.ArrayIndex(xloc, int(i))
				bld.GenTransfer(res, eltLoc)
			}
			p.vmap[i9n] = res

		}
	case *ssa.IndexAddr:

		switch i9n.X.Type().Underlying().(type) {
		case *types.Array:
			switch idx := i9n.Index.(type) {
			case *ssa.Const:
				i, ok := constant.Val(idx.Value).(int)
				if !ok {
					panic(fmt.Sprintf("const int %v\n", idx.Value))
				}
				xloc := p.vmap[i9n.X]
				p.vmap[i9n] = bld.ArrayIndex(xloc, i)
			default:
				// TBD
			}
		case *types.Slice:
			p.vmap[i9n] = p.vmap[i9n.X]
		default:
			// TBD
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
		bld.Type = i9n.Type()
		bld.Class = memory.Heap
		bld.SrcKind = results.SrcMakeSlice
		obj := bld.GenLoc()
		p.vmap[i9n] = obj
	case *ssa.MakeMap:
	case *ssa.MapUpdate:
	case *ssa.Next: // either string iterator or map
		if !i9n.IsString {
			// not addressable

			return nil
		}
		// it is a map, type Tuple
	case *ssa.Panic:
	case *ssa.Phi:
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
			// which don't have pointers....
			if vloc, ok := p.vmap[res]; ok {
				bld.GenTransfer(resLoc, vloc)
			}
		}

	case *ssa.UnOp:
		// Load

	case *ssa.Slice:
	case *ssa.Store:
		vloc := p.getLoc(bld, i9n.Val)
		aloc := p.getLoc(bld, i9n.Addr)
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

func (p *T) getLoc(b *results.Builder, v ssa.Value) memory.Loc {
	loc, ok := p.vmap[v]
	if ok {
		return loc
	}

	b.Reset()
	switch v.(type) {
	case *ssa.Global:
		b.Class = memory.Global
	default:
		b.Class = memory.Local
	}
	b.Pos = v.Pos()
	b.Type = v.Type().Underlying()
	loc = b.GenLoc()
	p.vmap[v] = loc
	if traceLocVal {
		fmt.Printf("i %s %s %p\n", v.Name(), plain.String(loc), v)
	}
	return loc
}

func (p *T) putResults() {
	if debugLogModel {
		fmt.Printf("built pal model for %s\n", p.pkgres.PkgPath)
		//p.pkgres.PlainEncode(os.Stdout)
	}
	p.results.Put(p.pass.Pkg.Path(), p.pkgres)
}
