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
)

func TestModel(t *testing.T) {
	mdl := NewModel(indexing.ConstVals())
	mdl.GenRoot(types.NewPointer(types.Typ[types.Int]), Global, IsParam|IsOpaque, token.NoPos)
	mdl.GenRoot(types.NewSlice(types.Typ[types.Float32]), Global, IsReturn, token.NoPos)
	gs := mdl.GenRoot(types.NewStruct([]*types.Var{
		types.NewVar(token.NoPos, nil, "f1", types.NewStruct([]*types.Var{
			types.NewVar(token.NoPos, nil, "f1i1", types.Typ[types.Int]),
			types.NewVar(token.NoPos, nil, "f1i2", types.Typ[types.Int])},
			[]string{"", ""})),
		types.NewVar(token.NoPos, nil, "f2", types.Typ[types.Int])},
		[]string{"", ""}), Global, 0, token.NoPos)

	gf := mdl.Field(gs, 1)
	if mdl.Parent(gf) != gs {
		t.Errorf("field parent")
	}
	if err := mdl.Check(); err != nil {
		t.Errorf("check failed: %s", err)
		return
	}

	//fmt.Printf(plain.String(mdl))
}
