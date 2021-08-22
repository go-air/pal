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
	"fmt"
	"io"
	"strconv"

	"github.com/go-air/pal/xtruth"
)

type consts struct{}
type C struct{ p *int64 }

func (c C) PlainEncode(w io.Writer) error {
	if c.p == nil {
		_, err := fmt.Fprintf(w, ".")
		return err
	}
	_, err := w.Write(strconv.AppendInt(nil, *c.p, 16))
	return err
}

func isHexLower(b byte) bool {
	return (b >= byte('0') && b <= byte('9')) || (b >= byte('a') && b <= byte('f'))
}

func (c C) PlainDecode(r io.Reader) error {
	var buf [16]byte
	_, err := io.ReadFull(r, buf[:1])
	if err != nil {
		return err
	}
	if buf[0] == byte('.') {
		c.p = nil
		return nil
	}
	i := 1
	for i < 16 {
		_, err = io.ReadFull(r, buf[i:i+1])
		if err != nil {
			return err
		}
		if !isHexLower(buf[i]) {
			break
		}

		i++
	}
	v, _ := strconv.ParseInt(string(buf[:i]), 16, 64)
	*c.p = v
	return nil
}

func ConstVals() T[C] {
	return consts{}
}

var zero int64 = 0
var one int64 = 1

func (c consts) Zero() C {
	z := int64(0)
	return C{&z}
}

func (c consts) One() C {
	o := int64(1)
	return C{&o}
}

func (c consts) IsVar(v C) bool {
	return v.p == nil
}

func (c consts) Var() C {
	return C{nil}
}

func (c consts) FromInt64(v int64) C {
	return C{&v}
}

func (c consts) ToInt64(v C) (int64, bool) {
	p := v.p
	if p == nil {
		return 0, false
	}
	return *p, true
}

func (c consts) Plus(a, b C) C {
	if a.p == nil || b.p == nil {
		return c.Var()
	}
	r := *a.p + *b.p
	return c.FromInt64(r)
}

func (c consts) Times(a, b C) C {
	pa, pb := a.p, b.p
	if pa == nil || pb == nil {
		return c.Var()
	}
	r := *pa * *pb
	return c.FromInt64(r)
}

func (c consts) Div(a, b C) (C, xtruth.T) {
	switch c.Equal(b, c.Zero()) {
	case xtruth.True:
		return c.Zero(), xtruth.False
	case xtruth.False:
		pa, pb := a.p, b.p
		r := *pa / *pb
		return c.FromInt64(r), xtruth.True
	case xtruth.X:
		return c.Var(), xtruth.X
	default:
		panic("xtruth")
	}
}

func (c consts) Rem(a, b C) (C, xtruth.T) {
	switch c.Equal(b, c.Zero()) {
	case xtruth.True:
		return c.Zero(), xtruth.False
	case xtruth.False:
		pa, pb := a.p, b.p
		r := *pa % *pb
		return c.FromInt64(r), xtruth.True
	case xtruth.X:
		return c.Var(), xtruth.X
	default:
		panic("xtruth")
	}
}

func (c consts) Band(a, b C) C {
	pa, pb := a.p, b.p
	if pa == nil || pb == nil {
		return c.Var()
	}
	z := *pa & *pb
	return c.FromInt64(z)
}

func (c consts) Bnot(a C) C {
	pa := a.p
	if pa == nil {
		return c.Var()
	}
	z := ^(*pa)
	return c.FromInt64(z)
}

func (c consts) Lshift(a, s C) (C, xtruth.T) {
	panic("unimplemented")
}

func (c consts) Rshift(a, s C) (C, xtruth.T) {
	panic("unimplemented")
}

func (c consts) Less(a, b C) xtruth.T {
	pa, pb := a.p, b.p
	if pa == nil || pb == nil {
		return xtruth.X
	}
	if *pa < *pb {
		return xtruth.True
	}
	return xtruth.False
}

func (c consts) Equal(a, b C) xtruth.T {
	pa, pb := a.p, b.p
	if pa == nil || pb == nil {
		return xtruth.X
	}
	if *pa == *pb {
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
