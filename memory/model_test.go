package memory

import (
	"fmt"
	"go/token"
	"go/types"
	"testing"

	"github.com/go-air/pal/internal/plain"
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

	fmt.Printf(plain.String(mdl))
}
