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
	if tracePackage {
		fmt.Printf("ssa2pal translating %s\n", p.pass.Pkg.Path())
	}
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
		_, ptr := p.buildr.GoType(ty.Elem().Underlying()).WithPointer()
		p.vmap[x] = ptr
	default:
		msg := fmt.Sprintf(
			"unexpected ssa global member type for %s %T\n",
			name,
			x.Type().Underlying())

		panic(msg)
	}
}

func (p *T) addFuncDecl(name string, fn *ssa.Function) error {
	if traceFunc {
		fmt.Printf("ssa2pal adding \"%s\".%s\n", p.pass.Pkg.Path(), fn.Name())
	}
	opaque := memory.NoAttrs
	if token.IsExported(name) {
		opaque = memory.IsOpaque
	}
	memFn := p.buildr.Func(fn.Signature, name, opaque)

	p.vmap[fn] = memFn.Loc()
	fmt.Printf("built func %s at %d\n", name, memFn.Loc())

	for i, param := range fn.Params {
		p.vmap[param] = p.buildr.Memory().Obj(memFn.ParamLoc(i))
		if traceParam {
			fmt.Printf("setting param %s to %d\n", param, p.vmap[param])
		}
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
	var ssaVal ssa.Value
	var ok bool
	for _, i9n := range blk.Instrs {
		ssaVal = nil
		switch v := i9n.(type) {
		case *ssa.DebugRef, *ssa.Defer, *ssa.Go, *ssa.If, *ssa.Jump,
			*ssa.MapUpdate, *ssa.Panic, *ssa.Return,
			*ssa.RunDefers, *ssa.Send, *ssa.Store:
			// these are not values
		default:
			ssaVal = v.(ssa.Value)
		}
		rands = i9n.Operands(rands[:0])
		for _, arg := range rands {
			_, ok := p.vmap[*arg]
			if !ok && *arg != nil {
				// I believe that if *arg is nil, the
				// arg is irrelevant (eg Slice with 1 or 2 args).
				p.genValueLoc(*arg)
			}
		}
		if ssaVal == nil {
			continue
		}
		_, ok = p.vmap[ssaVal]
		if !ok {
			p.genValueLoc(ssaVal)
		}
	}
}

// genValueLoc generates a memory.Loc associated with v.
//
// genValueLoc may need to work recursively on struct and
// array typed structured data.
func (p *T) genValueLoc(v ssa.Value) memory.Loc {
	if traceGenValueLoc {
		fmt.Printf("genValue for %s (%#v)\n", v, v)
	}
	switch v := v.(type) {
	case *ssa.Range, *ssa.Const:
		return memory.NoLoc
	default:
		p.buildr.Pos(v.Pos()).GoType(v.Type()).Class(memory.Local).Attrs(memory.NoAttrs)
	}
	var res memory.Loc
	switch v := v.(type) {
	case *ssa.Alloc:
		if v.Heap {
			p.buildr.Class(memory.Global)
		}
		p.buildr.GoType(v.Type().Underlying().(*types.Pointer).Elem())
		_, res = p.buildr.WithPointer()
	case *ssa.MakeSlice:
		res = p.buildr.Slice(v.Type().Underlying().(*types.Slice),
			p.indexing.Var(),
			p.indexing.Var()).Loc()
	case *ssa.MakeMap:
		res = p.buildr.Map(v.Type().Underlying().(*types.Map)).Loc()

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
		switch idx := v.Index.(type) {
		case *ssa.Const:
			i64, ok := constant.Int64Val(idx.Value)
			if !ok {
				panic("type checked const index not precise as int64")
			}
			i := int(i64) // should be ok also b/c it is type checked.
			res = x.At(i)
		default:
			if x.Len() == 0 {
				// this should be type checked, but it is
				// not.
				res = p.buildr.Memory().Zero()
			} else {
				// we have a variable or expression index.
				// we
				//  1. take address of the array, call it pa, in a new loc
				//  2. create qa, same type as pa
				//  3. add AddTransferIndex(qa, pa, p.indexing.Var())
				//  4. create res, type of element of array
				//  5. create res = load(qa)
				ty, ok := v.X.Type().Underlying().(*types.Array)
				if !ok {
					panic(fmt.Sprintf("v type %s %#v\n", v.Type(), v.Type()))
				}
				eltTy := ty.Elem()
				ptrTy := types.NewPointer(ty.Elem())
				pelt := p.buildr.GoType(ptrTy).Gen()
				p.buildr.AddAddressOf(pelt, x.At(0))
				qelt := p.buildr.Gen()
				// it may crash if oob, add address of nil
				// TBD: see if with indexing we can constrain this.
				p.buildr.AddAddressOf(qelt, p.buildr.Memory().Zero())
				res = p.buildr.GoType(eltTy).Gen()
				p.buildr.AddTransferIndex(qelt, pelt, p.indexing.Var())
				p.buildr.AddLoad(res, qelt)
			}
		}
	case *ssa.Extract:
		tloc := p.vmap[v.Tuple]
		if tloc == memory.NoLoc {
			tloc = p.genValueLoc(v.Tuple)
			// reset bld cfg for genLoc below
			p.buildr.GoType(v.Type())
		}
		tuple := p.buildr.Object(tloc).(*objects.Tuple)
		res = tuple.At(v.Index)

	case *ssa.Next:
		if v.IsString {
			tupty := types.NewTuple(
				types.NewVar(v.Pos(), nil, "#2", types.Typ[types.Bool]),
				types.NewVar(v.Pos(), nil, "#0", types.Typ[types.Int]),
				types.NewVar(v.Pos(), nil, "#1", types.Typ[types.Rune]))
			res = p.buildr.Tuple(tupty).Loc()
			p.vmap[v] = res
			return res
		}
		iter := v.Iter.(*ssa.Range)
		rxloc := p.vmap[iter.X]
		if rxloc == memory.NoLoc {
			rxloc = p.genValueLoc(iter.X)
			p.buildr.Pos(v.Pos()).Class(memory.Local).Attrs(memory.NoAttrs)
		}
		mgoty := iter.X.Type().Underlying().(*types.Map)
		m := p.buildr.Object(rxloc).(*objects.Map)
		tupty := types.NewTuple(
			types.NewVar(v.Pos(), nil, "#0", types.Typ[types.Bool]),
			types.NewVar(v.Pos(), nil, "#1", mgoty.Key()),
			types.NewVar(v.Pos(), nil, "#2", mgoty.Elem()))
		tuple := p.buildr.Tuple(tupty)
		res = tuple.Loc()
		p.buildr.AddTransfer(tuple.At(1), m.Key())
		p.buildr.AddTransfer(tuple.At(2), m.Elem())
	case *ssa.UnOp:
		xloc := p.vmap[v.X]
		if xloc == memory.NoLoc {
			xloc = p.genValueLoc(v.X)
			// reset bld cfg for genLoc below
			p.buildr.GoType(v.Type())
		}
		res = p.vmap[v]
		if res == memory.NoLoc {

			res = p.buildr.FromGoType(v.Type())
		}
		// Load, Recv<-
		switch v.Op {
		case token.MUL: // *p
			p.buildr.AddLoad(res, p.vmap[v.X])
		case token.ARROW:
			c := p.buildr.Object(p.vmap[v.X]).(*objects.Chan)
			c.Recv(res, p.buildr.Memory())

		default: // indexing
		}

	default:
		res = p.buildr.FromGoType(v.Type())

	}
	p.vmap[v] = res
	//fmt.Printf("genValueLoc(%s): %d\n", v, res)
	return res
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
	case *ssa.BinOp: // tbd: indexing
		switch i9n.Op {
		case token.ARROW:
			panic("send binop")
		default:

		}
	case *ssa.Call:
		p.call(i9n.Call, p.vmap[i9n])
	case *ssa.ChangeInterface:
	case *ssa.ChangeType:
	case *ssa.Convert:
	case *ssa.DebugRef:
	case *ssa.Defer:
		p.call(i9n.Call, memory.NoLoc)
	case *ssa.Extract: // done in gen locs
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
			p.buildr.AddAddressOf(out, fobj)
			mdl.SetObj(out, fobj)
		} else {
			mdl.AddTransferIndex(out, ptr, p.indexing.FromInt64(int64(i9n.Field)))
		}

	case *ssa.Go:
		// for now, treat as call
		p.call(i9n.Call, memory.NoLoc)
	case *ssa.If:
	case *ssa.Index: // constraints done in gen locs
	case *ssa.IndexAddr:
		ptr := p.vmap[i9n.X]
		res := p.vmap[i9n]
		switch i9n.X.Type().Underlying().(type) {
		case *types.Pointer: // to array
			p.buildr.AddTransferIndex(res, ptr, p.indexing.Var())
		case *types.Slice:
			p.buildr.AddTransferIndex(res, ptr, p.indexing.Var())
		default:
			panic("unexpected type of ssa.IndexAddr.X")
		}

	case *ssa.Jump: // no-op
	case *ssa.Lookup:
		obj := p.buildr.Object(p.vmap[i9n.X])
		switch ma := obj.(type) {
		case *objects.Map:
			ma.Lookup(p.vmap[i9n], p.buildr.Memory())
		default:
			// it is a string
		}
	case *ssa.MakeInterface: // constraints done in genLoc
	case *ssa.MakeClosure: // constraints done in genLoc
	case *ssa.MakeChan: // constraints done in genLoc
	case *ssa.MakeSlice: // constraints done in genLoc
	case *ssa.MakeMap: // constraints done in genLoc

	case *ssa.MapUpdate:
		mloc := p.vmap[i9n.Map]
		if mloc == memory.NoLoc {
			panic(fmt.Sprintf("no map %s %#v", i9n.Map, i9n.Map))
		}
		obj := p.buildr.Object(mloc)
		switch ma := obj.(type) {
		case *objects.Map:
			ma.Update(p.vmap[i9n.Key], p.vmap[i9n.Value], p.buildr.Memory())
		default:
			panic("huh?")
		}
	case *ssa.Next: // handled in genLoc
	case *ssa.Panic:
	case *ssa.Phi:
		v := p.vmap[i9n]
		for _, x := range i9n.Edges {
			ev := p.vmap[x]
			p.buildr.AddTransfer(v, ev)
		}
	case *ssa.Range: // everything is in ssa.Next, see genLoc
	case *ssa.RunDefers:
		// no-op b/c we treat defers like calls.
	case *ssa.Select: // all comm clauses handled via <- unary and binary ops.
	case *ssa.Send:
		c := p.buildr.Object(p.vmap[i9n.Chan]).(*objects.Chan)
		c.Send(p.vmap[i9n.X], p.buildr.Memory())
	case *ssa.Return:
		var ssaFn *ssa.Function = i9n.Parent()
		var palFn *objects.Func
		palFn = p.funcs[ssaFn]
		// copy results to palFn results...
		for i, res := range i9n.Results {
			resptr := palFn.ResultLoc(i)
			resobj := p.buildr.Memory().Obj(resptr)
			vloc := p.vmap[res]
			p.buildr.AddTransfer(resobj, vloc)
		}

	case *ssa.UnOp:

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

func (p *T) call(c ssa.CallCommon, dst memory.Loc) {
	if c.IsInvoke() {
		p.invoke(c)
		return
	}
	callee := c.StaticCallee()
	if false && callee == nil {
		fmt.Printf("dst %d not static %#v\n", dst, c)
		return
	}
	switch fssa := c.Value.(type) {
	case *ssa.Builtin:
	case *ssa.MakeClosure:
	default: // eg *Function (static call)
		// dynamic dispatch
		floc := p.vmap[fssa]
		if floc == memory.NoLoc {
			panic("wilma!")
		}
		fn, ok := p.buildr.Object(floc).(*objects.Func)
		if !ok {
			fmt.Printf(" could not call '%s' loc %d type %s\n", fssa.Name(), floc, p.buildr.TypeSet().String(p.buildr.Memory().Type(floc)))
			return
		}
		args := make([]memory.Loc, len(c.Args))
		for i, argVal := range c.Args {
			args[i] = p.vmap[argVal]
		}
		p.buildr.Call(fn, dst, args)
	}
}

func (p *T) invoke(c ssa.CallCommon) {
}

func (p *T) putResults() {
	if debugLogModel {
		fmt.Printf("built pal model for %s\n", p.pkgres.PkgPath)
		p.buildr.TypeSet().PlainEncode(os.Stdout)
		p.buildr.Memory().PlainEncode(os.Stdout)
		p.buildr.PlainEncodeObjects(os.Stdout)
	}
	p.results.Put(p.pass.Pkg.Path(), p.pkgres)
}
