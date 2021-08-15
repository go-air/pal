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

package indexing

import (
	"io"

	"github.com/go-air/pal/xtruth"
)

type consts struct{}

func ConstVals() T {
	return consts{}
}

func (c consts) Zero() I {
	return I(0)
}

func (c consts) One() I {
	return I(1)
}

func (c consts) IsVar(v I) bool {
	return false
}

func (c consts) Var() I {
	panic("constant variable requested.")
}

func (c consts) FromInt(v int) I {
	return v
}

func (c consts) ToInt(v I) (int, bool) {
	vv, ok := v.(int)
	if !ok {
		return 0, false
	}
	return vv, true
}

func (c consts) Plus(a, b I) I {
	aa, bb := a.(int), b.(int)
	return I(aa + bb)
}

func (c consts) Times(a, b I) I {
	aa, bb := a.(int), b.(int)
	return I(aa * bb)
}

func (c consts) Div(a, b I) (I, xtruth.T) {
	switch c.Equal(b, c.Zero()) {
	case xtruth.True:
		return c.Zero(), xtruth.False
	case xtruth.False:
		return I(a.(int) / b.(int)), xtruth.True
	case xtruth.X:
		panic("non-const op")
	}
	return c.Zero(), xtruth.X
}

func (c consts) Rem(a, b I) (I, xtruth.T) {
	switch c.Equal(b, c.Zero()) {
	case xtruth.True:
		return c.Zero(), xtruth.False
	case xtruth.False:
		return I(a.(int) % b.(int)), xtruth.True
	case xtruth.X:
		panic("non-const op")
	}
	return c.Zero(), xtruth.X
}

func (c consts) Band(a, b I) I {
	aa, bb := a.(int), b.(int)
	return I(aa & bb)
}

func (c consts) Bnot(a I) I {
	aa := a.(int)
	return I(^aa)
}

func (c consts) Lshift(a, s I) (I, xtruth.T) {
	panic("unimplemented")
}

func (c consts) Rshift(a, s I) (I, xtruth.T) {
	panic("unimplemented")
}

func (c consts) Less(a, b I) xtruth.T {
	aa, bb := a.(int), b.(int)
	if aa < bb {
		return xtruth.True
	}
	return xtruth.False
}

func (c consts) Equal(a, b I) xtruth.T {
	aa, bb := a.(int), b.(int)
	if aa == bb {
		return xtruth.True
	}
	return xtruth.False
}

func (c consts) PlainEncode(w io.Writer) error {
	return nil
}

func (c consts) PlainDecode(r io.Reader) error {
	return nil
}
