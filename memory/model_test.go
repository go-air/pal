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

package memory

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/go-air/pal/values"
)

func TestModel(t *testing.T) {
	mdl := NewModel(values.ConstVals())
	mdl.Global(types.NewPointer(types.Typ[types.Int]), IsParam|IsOpaque)
	mdl.Global(types.NewSlice(types.Typ[types.Float32]), IsReturn)
	gs := mdl.Global(types.NewStruct([]*types.Var{
		types.NewVar(token.NoPos, nil, "f1", types.NewStruct([]*types.Var{
			types.NewVar(token.NoPos, nil, "f1i1", types.Typ[types.Int]),
			types.NewVar(token.NoPos, nil, "f1i2", types.Typ[types.Int])},
			[]string{"", ""})),
		types.NewVar(token.NoPos, nil, "f2", types.Typ[types.Int])},
		[]string{"", ""}), 0)

	gf, e := mdl.AccessField(gs, 1)
	if e != nil {
		t.Errorf("access field 1: %s", e)
		return
	}
	if mdl.Parent(gf) != gs {
		t.Errorf("access field parent")
	}
	if err := mdl.Check(); err != nil {
		t.Errorf("check failed: %s", err)
		return
	}

	//fmt.Printf(plain.String(mdl))
}
