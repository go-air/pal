package values

import (
	"fmt"

	"go/types"
)

type Const int

type consts struct{}

func Consts() T {
	return consts{}
}

func (c consts) Zero() V {
	return V(Const(0))
}

func (c consts) One() V {
	return V(Const(1))
}

func (c consts) Var(v V) bool {
	return false
}

func (c consts) Kind(_ V) ValueKind {
	return ConstKind
}

func (c consts) Const(v V) (int, bool) {
	vv, ok := v.(Const)
	if !ok {
		return 0, false
	}
	return int(vv), true
}

func (c consts) FromType(ty types.Type) V {
	switch ty := ty.(type) {
	case *types.Basic:
		return c.One()
	case *types.Pointer:
		return c.One()
	case *types.Array:
		elem := c.FromType(ty.Elem()).(Const)
		return V(Const(ty.Len()) * elem)
	case *types.Map:
		k := c.FromType(ty.Key()).(int)
		v := c.FromType(ty.Elem()).(int)
		return Const(k + v)

	default:
		panic(fmt.Sprintf("%s: unexpected/unimplemented", ty))
	}
}

func (c consts) Plus(a, b V) V {
	aa, bb := a.(Const), b.(Const)
	return V(Const(aa + bb))
}

func (c consts) Less(a, b V) AbsTruth {
	aa, bb := a.(Const), b.(Const)
	if aa < bb {
		return True
	}
	return False
}

func (c consts) Equal(a, b V) AbsTruth {
	aa, bb := a.(Const), b.(Const)
	if aa == bb {
		return True
	}
	return False
}
