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

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/typeset"
)

func gp() {
	ts := typeset.New()
	gp := NewGenParams(ts)
	_ = gp
}

func TestModel(t *testing.T) {
	mdl := NewModel(indexing.ConstVals())
	ts := typeset.New()
	gp := NewGenParams(ts)
	var ty types.Type = types.NewPointer(types.Typ[types.Int])
	mdl.Gen(gp.GoType(ty))
	ty = types.NewSlice(types.Typ[types.Float64])
	mdl.Gen(gp.GoType(ty))
	ty = types.NewStruct([]*types.Var{
		types.NewVar(token.NoPos, nil, "f1", types.NewStruct([]*types.Var{
			types.NewVar(token.NoPos, nil, "f1i1", types.Typ[types.Int]),
			types.NewVar(token.NoPos, nil, "f1i2", types.Typ[types.Int])},
			[]string{"", ""})),
		types.NewVar(token.NoPos, nil, "f2", types.Typ[types.Int])},
		[]string{"", ""})
	gs := mdl.Gen(gp.GoType(ty))

	gf := mdl.Field(gs, 1)
	if mdl.Parent(gf) != gs {
		t.Errorf("field parent")
	}

	//fmt.Printf(plain.String(mdl))
}
