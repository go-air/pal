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
	"os"
	"reflect"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/results"
	"github.com/go-air/pal/ssa2pal"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
)

var flagSet = flag.NewFlagSet("pal", flag.ExitOnError)
var palVersion = flagSet.Bool("V", false, "print out pal version")

type resultType int

func (r *resultType) AFact() {}

// SSAAnalyzer produces an Analyzer which
// works on golang.org/x/tools/go/ssa form.
func SSAAnalyzer() *analysis.Analyzer {
	// generate a unique results object
	// for every analyzer.
	palRes, err := results.New()
	if err != nil {
		panic(err.Error())
	}
	return &analysis.Analyzer{
		Name:       "pal",
		Flags:      *flagSet,
		Doc:        doc, // see file paldoc.go
		Run:        run,
		Requires:   []*analysis.Analyzer{buildssa.Analyzer},
		ResultType: reflect.TypeOf(palRes),
		FactTypes:  []analysis.Fact{palRes}}
}

func run(pass *analysis.Pass) (interface{}, error) {
	//flagSet.Parse(pass.Analyzer.Flags.Args())
	if *palVersion {
		v, e := Version()
		if e != nil {
			return nil, e
		}
		fmt.Printf("%s\n", v)
		os.Exit(0)
	}
	pal, err := ssa2pal.New(pass, indexing.ConstVals())
	if err != nil {
		return nil, err
	}
	return pal.GenResult()
}
