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
	"github.com/go-air/pal/memory"
	"github.com/go-air/pal/objects"
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
	// map from ssa.Value to memory locs

	vmap   map[ssa.Value]memory.Loc
	buildr *objects.Builder

	funcs map[*ssa.Function]*objects.Func
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
		buildr:   objects.NewBuilder(pkgPath, vs),
		vmap:     make(map[ssa.Value]memory.Loc, 8192),

		funcs: make(map[*ssa.Function]*objects.Func)}
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
			p.genGlobal(name, g)
		}
	}

	// add funcs
	for _, name := range mbrKeys {
		mbr := mbrs[name]
		switch fn := mbr.(type) {
		case *ssa.Function:
			if err = p.addFuncDecl(name, fn); err != nil {
				return nil, err
			}
		}
	}
	// TBD: calc results from generation above

	// place the results for current package in p.results.
	p.putResults()
	return p.results, nil
}

func (p *T) genGlobal(name string, x *ssa.Global) {
	// globals are in general pointers to the globally stored
	// index
	p.buildr.Pos(x.Pos()).Class(memory.Global).Attrs(memory.NoAttrs)
	if token.IsExported(name) {
		// mark opaque because packages which import this one
		// may set the variable to whatever.
		p.buildr.Attrs(memory.IsOpaque)
	}
	switch ty := x.Type().Underlying().(type) {
	case *types.Pointer:
		p.vmap[x] = p.buildr.Pointer(ty).Loc()
	default:
		msg := fmt.Sprintf(
			"unexpected ssa global member type for %s %T\n",
			name,
			x.Type().Underlying())

		panic(msg)
	}
}

func (p *T) addFuncDecl(name string, fn *ssa.Function) error {
	opaque := memory.NoAttrs
	if token.IsExported(name) {
		opaque = memory.IsOpaque
	}
	memFn := p.buildr.Func(fn.Signature, name, opaque)

	p.vmap[fn] = memFn.Loc()

	for i, param := range fn.Params {
		p.vmap[param] = memFn.ParamLoc(i)
	}
	// free vars not needed here -- top level func def

	p.funcs[fn] = memFn
	p.genBlocksValues(name, fn)
	p.genConstraints(name, fn)

	return nil
}

func (p *T) genBlocksValues(name string, fn *ssa.Function) {
	for _, blk := range fn.Blocks {
		p.genBlockValues(name, blk)
	}
	if fn.Recover != nil {
		p.genBlockValues(name, fn.Recover)
	}
}

func (p *T) genConstraints(name string, fn *ssa.Function) {
	for _, blk := range fn.Blocks {
		p.genBlockConstraints(name, blk)
	}
	if fn.Recover != nil {
		p.genBlockConstraints(name, fn.Recover)
	}
}

func (p *T) genBlockValues(name string, blk *ssa.BasicBlock) {
	rands := make([]*ssa.Value, 0, 13)
	for _, i9n := range blk.Instrs {
		switch v := i9n.(type) {
		case *ssa.DebugRef, *ssa.Defer, *ssa.Go, *ssa.If, *ssa.Jump,
			*ssa.MapUpdate, *ssa.Panic, *ssa.Return,
			*ssa.RunDefers, *ssa.Send, *ssa.Store:
			// these are not values
		default:
			vv := v.(ssa.Value)
			_, ok := p.vmap[vv]
			if !ok {
				p.genValueLoc(v.(ssa.Value))
			}
		}
		rands = i9n.Operands(rands[:0])
		for _, arg := range rands {
			_, ok := p.vmap[*arg]
			if !ok && *arg != nil {
				// I believe that if *arg is nil, the
				// arg is irrelevant.  This does happen.
				p.genValueLoc(*arg)
			}
		}
	}
}

// genValueLoc generates a memory.Loc associated with v.
//
// genValueLoc may need to work recursively on struct and
// array typed structured data.
func (p *T) genValueLoc(v ssa.Value) memory.Loc {
	p.buildr.Pos(v.Pos()).GoType(v.Type()).Class(memory.Local).Attrs(memory.NoAttrs)
	var res memory.Loc
	switch v := v.(type) {
	case *ssa.Alloc:
		res = p.buildr.Gen()
	case *ssa.MakeSlice:
		res = p.buildr.Slice(v.Type().(*types.Slice),
			p.indexing.Var(),
			p.indexing.Var()).Loc()
	case *ssa.MakeMap:
		res = p.buildr.Map(v.Type().(*types.Map)).Loc()

	case *ssa.Field:
		xloc, ok := p.vmap[v.X]
		if !ok {
			xloc = p.genValueLoc(v.X)
			// reset bld cfg for genLoc below
			p.buildr.Pos(v.Pos()).GoType(v.Type()).Class(memory.Local).Attrs(memory.NoAttrs)
		}
		res = p.buildr.Object(xloc).(*objects.Struct).Field(v.Field)

	case *ssa.Index:
		xloc, ok := p.vmap[v.X]
		if !ok {
			xloc = p.genValueLoc(v.X)
			// reset bld cfg for genLoc below
			p.buildr.Pos(v.Pos()).GoType(v.Type()).Class(memory.Local).Attrs(memory.NoAttrs)
		}
		x := p.buildr.Object(xloc).(*objects.Array)
		switch v := v.Index.(type) {
		case *ssa.Const:
			i64, ok := constant.Int64Val(v.Value)
			if !ok {
				panic("type checked const index not precise as int64")
			}
			i := int(i64) // should be ok also b/c it is type checked.
			res = x.At(i)
		default:
			// we have a variable or expression
			// generate a new loc and transfer
			// all indices to it.
			// TBD: use indexing
			ty := v.Type().Underlying().(*types.Array)
			eltTy := ty.Elem()
			res = p.buildr.GoType(eltTy).Gen()
			N := int(ty.Len())
			for i := 0; i < N; i++ {
				eltLoc := x.At(i)
				p.buildr.AddTransfer(res, eltLoc)
			}
		}

	default:
		res = p.buildr.Gen()
	}
	p.vmap[v] = res
	return res
}

// genAlloc generates a memory.Loc associated with an *ssa.Alloc
// we handle these specially because they generate the allocated
// object but are associated with pointers to these objects.
func (p *T) genAlloc(a *ssa.Alloc) memory.Loc {
	if a.Heap {
		p.buildr.Class(memory.Global)
	}
	p.buildr.GoType(a.Type().(*types.Pointer).Elem())
	p.buildr.Pos(a.Pos())

	_, ptr := p.buildr.WithPointer()
	return ptr
}

// generate all constraints for blk
// value nodes already have memory.Locs
func (p *T) genBlockConstraints(fnName string, blk *ssa.BasicBlock) error {

	for _, i9n := range blk.Instrs {
		if err := p.genI9nConstraints(fnName, i9n); err != nil {
			return err
		}
	}
	return nil
}

func (p *T) genI9nConstraints(fnName string, i9n ssa.Instruction) error {
	if traceGenI9n {
		fmt.Printf("gen %s\n", i9n)
	}
	switch i9n := i9n.(type) {
	case *ssa.Alloc: // done in gen locs
	case *ssa.BinOp:
	case *ssa.Call:
		p.call(i9n.Call)
	case *ssa.ChangeInterface:
	case *ssa.ChangeType:
	case *ssa.Convert:
	case *ssa.DebugRef:
	case *ssa.Defer:
		p.call(i9n.Call)
	case *ssa.Extract:
	case *ssa.Field: // done in gen locs

	case *ssa.FieldAddr:
		// i9n.X is a pointer to struct
		// the result is the address of the field
		// indicated by i9n.Field (an int)

		ptr := p.vmap[i9n.X]
		out := p.vmap[i9n]

		mdl := p.buildr.Memory()
		obj := mdl.Obj(ptr)
		fobj := memory.NoLoc
		if obj != memory.NoLoc {
			fobj = mdl.Field(obj, i9n.Field)
			p.buildr.AddPointsTo(out, fobj)
			mdl.SetObj(out, fobj)
		} else {
			mdl.AddTransferIndex(out, ptr, p.indexing.FromInt64(int64(i9n.Field)))
		}

	case *ssa.Go:
		// for now, treat as call
		p.call(i9n.Call)
	case *ssa.If:
	case *ssa.Index: // constraints done in gen locs
	case *ssa.IndexAddr:
		ptr := p.vmap[i9n.X]
		res := p.vmap[i9n]
		switch i9n.X.Type().Underlying().(type) {
		case *types.Pointer: // to array
			p.buildr.AddTransfer(res, ptr)
		case *types.Slice:
			p.buildr.AddTransfer(res, ptr)
		default:
			panic("unexpected type of ssa.IndexAddr.X")
		}

	case *ssa.Jump:
	case *ssa.Lookup:
	case *ssa.MakeInterface:
	case *ssa.MakeClosure:
	case *ssa.MakeChan:
	case *ssa.MakeSlice: // constraints done in genLoc
	case *ssa.MakeMap: // constraints done in genLoc

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
			p.buildr.AddTransfer(v, ev)
		}
	case *ssa.Range:
	case *ssa.RunDefers:
		// no-op b/c we treat defers like calls.
	case *ssa.Select:
	case *ssa.Send:
	case *ssa.Return:
		var ssaFn *ssa.Function = i9n.Parent()
		var palFn *objects.Func
		palFn = p.funcs[ssaFn]
		// copy results to palFn results...
		for i, res := range i9n.Results {
			resLoc := palFn.ResultLoc(i)
			vloc := p.vmap[res]
			p.buildr.AddTransfer(resLoc, vloc)
		}

	case *ssa.UnOp:
		// Load
		switch i9n.Op {
		case token.MUL: // *p
			p.buildr.AddLoad(p.vmap[i9n], p.vmap[i9n.X])
		case token.ARROW: // <- TBD:

		default:
		}

	case *ssa.Slice:
		p.buildr.AddTransfer(p.vmap[i9n], p.vmap[i9n.X])
	case *ssa.Store:
		vloc := p.vmap[i9n.Val]
		aloc := p.vmap[i9n.Addr]
		p.buildr.AddStore(aloc, vloc)

	case *ssa.TypeAssert:
	default:
		panic("unknown ssa Instruction")
	}
	return nil
}

func (p *T) PkgPath() string {
	return p.pass.Pkg.Path()
}

func (p *T) call(c ssa.CallCommon) {
	if c.IsInvoke() {
		p.invoke(c)
		return
	}
	callee := c.StaticCallee()
	if callee == nil {
		return
	}

}

func (p *T) invoke(c ssa.CallCommon) {
}

func (p *T) putResults() {
	if debugLogModel {
		fmt.Printf("built pal model for %s\n", p.pkgres.PkgPath)
		p.pkgres.PlainEncode(os.Stdout)
	}
	p.results.Put(p.pass.Pkg.Path(), p.pkgres)
}
