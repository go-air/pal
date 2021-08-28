package objects

import (
	"bytes"
	"go/token"
	"go/types"
	"testing"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/memory"
)

func TestBuilder(t *testing.T) {
	cvs := indexing.ConstVals()
	br := NewBuilder("", cvs)
	br.Array(types.NewArray(types.Typ[types.Int], 11))
	t1 := types.NewTuple(
		types.NewVar(token.NoPos, nil, "v1", types.Typ[types.Bool]),
		types.NewVar(token.NoPos, types.NewPackage("a/b", "b"), "v2", types.Typ[types.String]))
	t2 := types.NewTuple()
	sig := types.NewSignature(nil, t1, t2, false)
	f := br.Func(sig, "fname", memory.IsOpaque)
	_ = f
	io := bytes.NewBuffer(nil)
	if err := br.PlainEncodeObjects(io); err != nil {
		t.Error(err)
	}
}
