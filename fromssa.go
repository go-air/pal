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
	"fmt"

	"github.com/go-air/pal/results"
	"github.com/go-air/pal/values"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

type FromSSA struct {
	pass *analysis.Pass
	ssa  *buildssa.SSA
	// this is a state variable which
	// represents the current package under
	// analysis.
	pkg     *ssa.Package
	values  values.T
	results *results.T
}

var noImp = map[string]struct{}{
	"unsafe":      struct{}{},
	"runtime":     struct{}{},
	"runtime/cpu": struct{}{},
	"runtype/sys": struct{}{}}

func NewFromSSA(pass *analysis.Pass, vs values.T) (*FromSSA, error) {
	var palres *results.T
	var err error
	for _, imp := range pass.Pkg.Imports() {
		_, present := noImp[imp.Path()]
		if present {
			continue
		}
		if !imp.Complete() {
			return nil, fmt.Errorf("%s incomplete", imp.Name())
		}
		if palres != nil {
			continue
		}
		if !pass.ImportPackageFact(imp, palres) {
			return nil, fmt.Errorf("unable to import from %s", imp.Name())
		}
	}
	if palres == nil {
		palres, err = results.NewT()
		if err != nil {
			return nil, err
		}
	}
	ssa := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	ssa.Pkg.Build()
	fromSSA := &FromSSA{
		pass:    pass,
		pkg:     ssa.Pkg,
		ssa:     ssa,
		results: palres,
		values:  vs}
	return fromSSA, nil
}

/*
func (f *FromSSA) registerAlloc(a *ssa.Alloc) mem.T {
	if !a.Heap {
		return mem.T(0)
	}
	var m mem.T
	m = f.mems.Heap(a.Type(), 0)
	f.set(m, &i)
	return m
}

func (f *FromSSA) registerGlobal(g *ssa.Global) mem.T {
	var m mem.T
	var i = MemSSA{Pkg: f.pkg, Global: g}
	m = f.mems.Global(g.Type(), 0)
	f.set(m, &i)
	return m
}
*/
