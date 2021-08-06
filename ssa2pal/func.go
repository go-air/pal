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
	fnobj    memory.Loc
	fnloc    memory.Loc
	recv     memory.Loc
	params   []memory.Loc
	results  []memory.Loc
}

func tupleLen(tuple *types.Tuple) int {
	if tuple == nil {
		return 0
	}
	return tuple.Len()
}

func NewFunc(bld *results.Builder, sig *types.Signature, declName string) *Func {
	fn := &Func{sig: sig, declName: declName}
	bld.Reset()
	bld.Class = memory.Global
	bld.SrcKind = results.SrcFunc
	bld.Pos = token.NoPos // XXX
	bld.Type = sig
	fn.fnobj = bld.GenLoc()
	fn.fnloc = bld.GenPointerTo(fn.fnobj)
	fn.params = make([]memory.Loc, tupleLen(sig.Params()))
	fn.results = make([]memory.Loc, tupleLen(sig.Results()))

	// handle parameters
	opaque := memory.NoAttrs
	if token.IsExported(declName) {
		opaque |= memory.IsOpaque
	}
	bld.Class = memory.Local // always
	bld.SrcKind = results.SrcVar

	recv := sig.Recv()
	if recv != nil {
		bld.Class = memory.Local
		fn.recv = bld.GenLoc()
	}
	params := sig.Params()
	if params != nil {
		N := params.Len()
		for i := 0; i < N; i++ {
			param := params.At(i)
			bld.Pos = param.Pos()
			bld.Type = param.Type()
			bld.Attrs = memory.IsParam | opaque
			fn.params[i] = bld.GenLoc()
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
			fn.results[i] = bld.GenLoc()
		}
	}
	// TBD: FreeVars
	return fn
}

func (f *Func) Declared() bool {
	return f.declName != ""
}

func (f *Func) Name() string {
	return f.declName
}

func (f *Func) Loc() memory.Loc {
	return f.fnloc
}

func (f *Func) Obj() memory.Loc {
	return f.fnobj
}

func (f *Func) Sig() *types.Signature {
	return f.sig
}

func (f *Func) RecvLoc(i int) memory.Loc {
	return f.recv

}

func (f *Func) ParamLoc(i int) memory.Loc {
	return f.params[i]
}

func (f *Func) NParams() int {
	return len(f.params)
}

func (f *Func) ResultLoc(i int) memory.Loc {
	return f.results[i]
}

func (f *Func) NResults() int {
	return len(f.results)
}
