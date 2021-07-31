package values

import (
	"fmt"
	"go/types"
)

// TypeVSize gives a virtual sizeof a type.
// In pal, memory locations are always type
// aligned, so the VSize is just a count of
// the number of basic types, indicating the
// size, in terms of number of locations,
// of data represented by a type.
func TypeVSize(ty types.Type) int {
	switch ty := ty.(type) {
	case *types.Basic:
		return 1
	case *types.Pointer:
		return 1
	case *types.Array:
		elem := TypeVSize(ty.Elem())
		return int(ty.Len()) * elem
	case *types.Map:
		k := TypeVSize(ty.Key())
		v := TypeVSize(ty.Elem())
		return k + v
	case *types.Struct:
		n := ty.NumFields()
		sum := 0
		for i := 0; i < n; i++ {
			fty := ty.Field(i).Type()
			sum += TypeVSize(fty)
		}
		return sum
	case *types.Named:
		return TypeVSize(ty.Underlying())

	default:
		panic(fmt.Sprintf("%s: unexpected/unimplemented", ty))
	}
}
