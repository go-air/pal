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
	"flag"
	"fmt"
	"reflect"

	"github.com/go-air/pal/mem"
	"github.com/go-air/pal/values"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

var flagSet = flag.NewFlagSet("pal", flag.ExitOnError)

func Analyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:       "pal",
		Flags:      *flagSet,
		Doc:        "pal pointer analysis",
		Run:        run,
		Requires:   []*analysis.Analyzer{buildssa.Analyzer},
		ResultType: reflect.TypeOf((*mem.Model)(nil)),
		FactTypes:  []analysis.Fact{&PkgFact{}}}
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, imp := range pass.Pkg.Imports() {
		if !imp.Complete() {
			return nil, fmt.Errorf("%s incomplete", imp.Name())
		}
		f := &PkgFact{}
		if !pass.ImportPackageFact(imp, f) {
			return nil, fmt.Errorf("unable to import from %s", imp.Name())
		}
	}
	ssa := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	fmt.Printf("building pkg %s\n", ssa.Pkg.Pkg.Name())
	ssa.Pkg.Build()
	fromSSA := NewFromSSA(ssa, values.Consts())
	_ = fromSSA
	mems := mem.NewModel(values.Consts())
	// bind globals

	// bind funcs
	for _, m := range ssa.Pkg.Members {
		runMember(ssa, m, mems)
	}
	for _, fn := range ssa.SrcFuncs {
		runFunc(ssa, fn, mems)

	}
	return new(mem.Model), nil
}

func runMember(_ *buildssa.SSA, m ssa.Member, mems *mem.Model) {
	switch m := m.(type) {
	case *ssa.Global:
		fmt.Printf("global %s\n", m)
		mems.Global(m.Type(), 0)
	default:
	}
}

func runFunc(ssa *buildssa.SSA, fn *ssa.Function, mems *mem.Model) {
	// bind params
	for _, p := range fn.Params {
		_ = p
		//fmt.Printf("%s - %s\n", p.Name(), p.Object())
	}
	for _, b := range fn.Blocks {
		_ = b
		runBlock(b, mems)

	}
}

func runBlock(block *ssa.BasicBlock, mems *mem.Model) {
	for _, i9n := range block.Instrs {
		runOne(i9n, mems)
	}
}

func runOne(n ssa.Instruction, mems *mem.Model) {
	rands := make([]ssa.Value, 0, 128)
	_ = rands
	switch n := n.(type) {
	default:
		_ = n
		//fmt.Printf("\t%#v %s %T\n", n, n, n)
	}
	switch n := n.(type) {
	case *ssa.Alloc:
		if n.Heap {
		}
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
	case *ssa.Go:
	case *ssa.If:
	case *ssa.Index:
	case *ssa.IndexAddr:
	case *ssa.Jump:
	case *ssa.Lookup:
	case *ssa.MakeInterface:
	case *ssa.MakeClosure:
	case *ssa.MakeChan:
	case *ssa.MakeSlice:
	case *ssa.MakeMap:
	case *ssa.MapUpdate:
	case *ssa.Next:
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
