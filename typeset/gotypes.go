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

package typeset

import (
	"fmt"
	"go/types"
)

func (t *TypeSet) FromGoType(gotype types.Type) Type {
	return t.fromGoType(gotype, make(map[string]bool), false)
}

func (t *TypeSet) fromGoType(gotype types.Type, seen map[string]bool, ignoreRecv bool) Type {
	// that line is a headache...
	switch ty := gotype.(type) {
	case *types.Basic:
		switch ty.Kind() {
		case types.Bool:
			return Bool
		case types.Int:
			return Int64
		case types.Uint:
			return Uint64

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
			// XXX
			// this can be untyped experimentally
			// but then it is constant, and so
			// not address taken, and so irrelevant
			// to pointer analysis.  Just give it
			// a type and continue.
			return Bool
			//panic(fmt.Sprintf("%s has no pal type", ty))
		}

	case *types.Slice:
		elty := t.fromGoType(ty.Elem(), seen, false)
		return t.getSlice(elty)
	case *types.Pointer:
		elty := t.fromGoType(ty.Elem(), seen, false)
		return t.getPointer(elty)
	case *types.Chan:
		elty := t.fromGoType(ty.Elem(), seen, false)
		return t.getChan(elty)
	case *types.Array:
		elty := t.fromGoType(ty.Elem(), seen, false)
		// TBD: consider int v int64
		return t.getArray(elty, int(ty.Len()))

	case *types.Struct:
		N := ty.NumFields()
		fields := make([]named, N)

		off := 1
		for i := 0; i < N; i++ {
			goField := ty.Field(i)
			fty := t.fromGoType(goField.Type(), seen, false)
			fields[i] = named{name: goField.Name(), typ: fty, loff: off}
			off += t.Lsize(fty)
		}
		return t.getStruct(fields)

	case *types.Map:
		kty := t.fromGoType(ty.Key(), seen, false)
		ety := t.fromGoType(ty.Elem(), seen, false)
		return t.getMap(kty, ety)

	case *types.Interface:
		ty = ty.Complete()
		N := ty.NumMethods()
		fields := make([]named, N)
		for i := 0; i < N; i++ {
			meth := ty.Method(i)
			mty := t.fromGoType(meth.Type(), seen, true)
			mname := meth.Name()
			fields[i] = named{name: mname, typ: mty}
		}
		return t.getInterface(fields)
	case *types.Signature:
		goParams := ty.Params()
		goResults := ty.Results()
		params := make([]named, goParams.Len())
		results := make([]named, goResults.Len())
		for i := range params {
			param := goParams.At(i)
			params[i].name = param.Name()
			params[i].typ = t.fromGoType(param.Type(), seen, false)
		}
		for i := range results {
			result := goResults.At(i)
			results[i].name = result.Name()
			results[i].typ = t.fromGoType(result.Type(), seen, false)
		}
		recv := NoType
		if !ignoreRecv && ty.Recv() != nil {
			recv = t.fromGoType(ty.Recv().Type(), seen, false)
		}
		return t.getSignature(recv, params, results, ty.Variadic())

	case *types.Tuple:
		N := ty.Len()
		res := make([]named, N)
		off := 0
		for i := 0; i < N; i++ {
			at := ty.At(i)
			res[i].name = at.Name()
			res[i].typ = t.fromGoType(at.Type(), seen, false)
			res[i].loff = off
			off += t.Lsize(res[i].typ)
		}
		return t.getTuple(res)

	case *types.Named:
		if !seen[ty.String()] {
			seen[ty.String()] = true
			return t.fromGoType(ty.Underlying(), seen, false)
		}
		return Type(0)
	default:
		panic(fmt.Sprintf("pal type cannot represent go type %s", gotype))
	}
}

func (t *TypeSet) ToGoType(ty Type) types.Type {
	return nil
}
