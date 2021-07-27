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

	"golang.org/x/tools/0.20210713164023-8e85a283cd8f/go/ssa"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

var flagSet = flag.NewFlagSet("pal", flag.ExitOnError)

var Analyzer = &analysis.Analyzer{
	Name:       "pal",
	Flags:      *flagSet,
	Doc:        "pal pointer analysis",
	Run:        run,
	Requires:   []*analysis.Analyzer{buildssa.Analyzer},
	ResultType: reflect.TypeOf((*Mems)(nil))}

func run(pass *analysis.Pass) (interface{}, error) {
	ssa := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	ssa.Pkg.Build()
	mems := &mems{}
	for _, fn := range ssa.SrcFuncs {
		runFunc(ssa, fn, mems)

	}
	return new(Mems), nil
}

func runFunc(ssa *buildssa.SSA, fn *ssa.Function, mems *mems) {
	// bind params
	for _, p := range fn.Params {
		fmt.Printf("%s - %s\n", p.Name(), p.Object())
	}
	for _, b := range fn.Blocks {


	}
}

func runBlock(block *ssa.BasicBlock, mems *mems) {
	rands := make([]ssa.Value, 0, 128)
	for _, i9n := range block.Instrs {
		runOne(i9n, mems)
	}
}

func runOne(n ssa.Node) {
		switch n := n.(type) {
		case *ssa.Alloc:
			if n.Heap {
			}
		case *ssa.Builtin:
		case *ssa.Binop:
		case *ssa.Call:
		case *ssa.ChangeInterface:
		case *ssa.ChangeType:
		case *ssa.Const:
		case *ssa.Convert:
		case *ssa.DebugRef:
		case *ssa.Defer:
		case *ssa.Extract:
		case *ssa.Field:
		case *ssa.FieldAddr:
		case *ssa.Function:
		case *ssa.Global:
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
		case *ssa.NamedConst:
		case *ssa.Next:
		case *ssa.Parameter:
		case *ssa.Panic:
		case *ssa.Phi:
		case *ssa.Select:
		case *ssa.Send:
		case *ssa.Return:
		case *ssa.UnOp:
		case *ssa.Slice:
		case *ssa.Store:
		case *ssa.Type:
		case *ssa.TypeAssert:
		default:
			panic("unknown ssa Node")
		}
	}
}
