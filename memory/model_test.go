package memory

import (
	"fmt"
	"go/types"
	"testing"

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/values"
)

func TestModel(t *testing.T) {
	mdl := NewModel(values.ConstVals())
	mdl.Global(types.NewPointer(types.Typ[types.Int]), IsParam|IsOpaque)
	mdl.Global(types.NewSlice(types.Typ[types.Float32]), IsReturn)
	fmt.Printf(plain.String(mdl))
}
