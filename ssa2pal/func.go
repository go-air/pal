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

package ssa2pal

import (
	"go/token"
	"go/types"

	"github.com/go-air/pal/memory"
	"github.com/go-air/pal/results"
)

type Func struct {
	sig      *types.Signature
	declName string
	locs     []memory.Loc
}

func NewFunc(bld *results.Builder, sig *types.Signature, declName string) *Func {
	bld.Reset()
	bld.Class = memory.Global
	bld.SrcKind = results.SrcFunc
	bld.Pos = token.NoPos // XXX
	bld.Type = sig
	locs := make([]memory.Loc, 1, 5)
	locs[0] = bld.GenLoc()

	// handle parameters
	opaque := memory.Attrs(0)
	if token.IsExported(declName) {
		opaque |= memory.IsOpaque
	}
	bld.Class = memory.Local // always
	bld.SrcKind = results.SrcVar

	recv := sig.Recv()
	if recv != nil {
		bld.Class = memory.Local
		bld.GenLoc()
	}
	params := sig.Params()
	if params != nil {
		N := params.Len()
		for i := 0; i < N; i++ {
			param := params.At(i)
			bld.Pos = param.Pos()
			bld.Type = param.Type()
			bld.Attrs = memory.IsParam | opaque
			locs = append(locs, bld.GenLoc())
		}
	}
	rets := sig.Results()
	if rets != nil {
		N := rets.Len()
		for i := 0; i < N; i++ {
			ret := rets.At(i)
			bld.Pos = ret.Pos()
			bld.Type = ret.Type()
			bld.Attrs = memory.IsReturn | opaque
			locs = append(locs, bld.GenLoc())
		}
	}
	res := &Func{locs: locs, sig: sig, declName: declName}
	return res
}

func (f *Func) Declared() bool {
	return f.declName != ""
}

func (f *Func) Name() string {
	return f.declName
}

func (f *Func) Loc() memory.Loc {
	return f.locs[0]
}

func (f *Func) Sig() *types.Signature {
	return f.sig
}

func (f *Func) RecvLoc(i int) memory.Loc {
	if f.sig.Recv() != nil {
		return f.locs[i+1]
	}
	panic("oob")
}

func (f *Func) ParamLoc(i int) memory.Loc {
	if f.sig.Recv() != nil {
		i++
	}
	parms := f.sig.Params()
	if parms != nil && i >= parms.Len() {
		panic("oob")
	}
	return f.locs[i]
}

func (f *Func) ResultLoc(i int) memory.Loc {
	if f.sig.Recv() != nil {
		i++
	}
	parms := f.sig.Params()
	if parms != nil {
		i += parms.Len()
	}
	return f.locs[i]
}
