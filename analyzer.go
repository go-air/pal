// Copyright 2021 Scott Cotton
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

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
)

var flagSet = flag.NewFlagSet("pal", flag.ExitOnError)

var Analyzer = &analysis.Analyzer{
	Name:       "pal",
	Flags:      *flagSet,
	Doc:        "pal pointer analysis",
	Run:        run,
	Requires:   []*analysis.Analyzer{buildssa.Analyzer},
	ResultType: reflect.TypeOf(new(Mems))}

func run(pass *analysis.Pass) (interface{}, error) {
	ssa := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	ssa.Pkg.Build()
	fmt.Printf("%s\n", ssa)
	return new(Mems), nil
}
