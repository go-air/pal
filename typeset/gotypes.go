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
		// TBD: consider int v int64
		return t.getArray(elty, int(ty.Len()))

	case *types.Struct:
		N := ty.NumFields()
		fields := make([]named, N)

		off := 1
		for i := 0; i < N; i++ {
			goField := ty.Field(i)
			fty := t.FromGoType(goField.Type())
			fields[i] = named{name: goField.Name(), typ: fty, loff: off}
			off += t.Lsize(fty)
		}
		return t.getStruct(fields)

	case *types.Map:
		kty := t.FromGoType(ty.Key())
		ety := t.FromGoType(ty.Elem())
		return t.getMap(kty, ety)

	case *types.Interface:
		ty = ty.Complete()
		N := ty.NumMethods()
		fields := make([]named, N)
		for i := 0; i < N; i++ {
			meth := ty.Method(i)
			mty := t.FromGoType(meth.Type())
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
			params[i].typ = t.FromGoType(param.Type())
		}
		for i := range results {
			result := goResults.At(i)
			results[i].name = result.Name()
			results[i].typ = t.FromGoType(result.Type())
		}

		var recv Type
		gorecv := ty.Recv()
		if gorecv != nil {
			recv = t.FromGoType(gorecv.Type())
		}
		return t.getSignature(recv, params, results, ty.Variadic())

	case *types.Tuple:
		N := ty.Len()
		res := make([]named, N)
		off := 1
		for i := 0; i < N; i++ {
			at := ty.At(i)
			res[i].name = at.Name()
			res[i].typ = t.FromGoType(at.Type())
			res[i].loff = off
			off += t.Lsize(res[i].typ)
		}
		return t.getTuple(res)

	case *types.Named:
		name := ty.String()
		res, creat := t.getNamed(name)
		if creat {
			under := t.FromGoType(ty.Underlying())
			t.nodes[res].elem = under
			t.nodes[res].lsize = t.Lsize(under)
		}
		return res
	default:
		panic(fmt.Sprintf("pal type cannot represent go type %s (%#v)", gotype, gotype))
	}
}

func (t *TypeSet) ToGoType(ty Type) types.Type {
	return nil
}
