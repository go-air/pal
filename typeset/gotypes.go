package typeset

import (
	"fmt"
	"go/types"
)

func (t *T) FromGoType(gotype types.Type) Type {
	// that line is a headache...
	switch ty := gotype.(type) {
	case *types.Basic:
		switch ty.Kind() {
		case types.Bool:
			return Bool
		case types.Int:
			return Int64

		case types.Int8:
			return Int8
		case types.Int16:
			return Int16
		case types.Int32:
			return Int32
		case types.Int64:
			return Int64
		case types.Uint8:
			return Uint8
		case types.Uint16:
			return Uint16
		case types.Uint32:
			return Uint32
		case types.Uint64:
			return Uint64
		case types.Float32:
			return Float32
		case types.Float64:
			return Float64
		case types.String:
			return String
		case types.Complex64:
			return Complex64
		case types.Complex128:
			return Complex128
		case types.UnsafePointer:
			return UnsafePointer
		case types.Uintptr:
			return Uintptr
		default:
			panic("untyped const has no pal type")
		}

	case *types.Slice:
		elty := t.FromGoType(ty.Elem())
		return t.getSlice(elty)
	case *types.Pointer:
		elty := t.FromGoType(ty.Elem())
		return t.getPointer(elty)
	case *types.Chan:
		elty := t.FromGoType(ty.Elem())
		return t.getChan(elty)
	case *types.Array:
		elty := t.FromGoType(ty.Elem())
		return t.getArray(elty, ty.Len())

	case *types.Struct:
	case *types.Map:
		elty := t.FromGoType(ty.Elem())
		keyty := t.FromGoType(ty.Key())
	case *types.Interface:
	case *types.Signature:
	case *types.Tuple:

	case *types.Named:
		return t.FromGoType(ty.Underlying())
	default:
		panic(fmt.Sprintf("pal type cannot represent go type %s", gotype))
	}
	return 0
}

func (t *T) ToGoType(ty Type) types.Type {
	return nil
}
