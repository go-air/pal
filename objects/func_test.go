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

package objects

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/memory"
)

func fclob(o Object) {
	f := o.(*Func)
	f.declName = ""
	f.fnobj = 0
	f.recv = 0
	f.variadic = false
	f.params = nil
	f.results = nil
}

func TestFuncPlain(t *testing.T) {
	parms := types.NewTuple(
		types.NewVar(token.NoPos, nil, "p1", types.Typ[types.Int]),
		types.NewVar(token.NoPos, nil, "p2", types.Typ[types.Bool]))
	ress := types.NewTuple(
		types.NewVar(token.NoPos, nil, "r1", types.Typ[types.Int8]))
	sig := types.NewSignature(nil, parms, ress, false)
	b := NewBuilder("pkg", indexing.ConstVals())
	f := b.Func(sig, "fname", memory.IsOpaque)
	err := testRoundTrip(f, fclob, true)
	if err != nil {
		t.Error(err)
	}
}
