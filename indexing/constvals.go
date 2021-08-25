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

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/xtruth"
)

type consts struct{}
type C struct{ p *int64 }

func (c C) PlainEncode(w io.Writer) error {
	if c.p == nil {
		_, err := fmt.Fprintf(w, ".")
		return err
	}
	err := plain.Put(w, "c")
	if err != nil {
		return err
	}
	return plain.EncodeInt64(w, *c.p)
}

func (c *C) PlainDecode(r io.Reader) error {
	var buf [16]byte
	_, err := io.ReadFull(r, buf[:1])
	if err != nil {
		return err
	}
	if buf[0] == byte('.') {
		c.p = nil
		return nil
	}
	if buf[0] != byte('c') {
		return fmt.Errorf("unexpected %c != c", buf[0])
	}
	var v int64
	v, err = plain.DecodeInt64From(r)
	if err != nil {
		return err
	}
	c.p = &v
	return nil
}

func (c *C) Gen() I {
	return &C{}
}

func ConstVals() T {
	return consts{}
}

var zero int64 = 0
var one int64 = 1

func (c consts) Zero() I {
	z := int64(0)
	return I(&C{&z})
}

func (c consts) One() I {
	o := int64(1)
	return I(&C{&o})
}

func (c consts) IsVar(v I) bool {
	return v.(*C).p == nil
}

func (c consts) Var() I {
	return &C{nil}
}

func (c consts) FromInt64(v int64) I {
	return &C{&v}
}

func (c consts) ToInt64(v I) (int64, bool) {
	p := v.(*C).p
	if p == nil {
		return 0, false
	}
	return *p, true
}

func (c consts) Plus(a, b I) I {
	pa, pb := a.(*C).p, b.(*C).p
	if pa == nil || pb == nil {
		return c.Var()
	}
	r := *pa + *pb
	return c.FromInt64(r)
}

func (c consts) Times(a, b I) I {
	pa, pb := a.(*C).p, b.(*C).p
	if pa == nil || pb == nil {
		return c.Var()
	}
	r := *pa * *pb
	return c.FromInt64(r)
}

func (c consts) Div(a, b I) (I, xtruth.T) {
	switch c.Equal(b, c.Zero()) {
	case xtruth.True:
		return c.Zero(), xtruth.False
	case xtruth.False:
		pa, pb := a.(*C).p, b.(*C).p
		r := *pa / *pb
		return c.FromInt64(r), xtruth.True
	case xtruth.X:
		return c.Var(), xtruth.X
	}
	return c.Zero(), xtruth.X
}

func (c consts) Rem(a, b I) (I, xtruth.T) {
	switch c.Equal(b, c.Zero()) {
	case xtruth.True:
		return c.Zero(), xtruth.False
	case xtruth.False:
		pa, pb := a.(*C).p, b.(*C).p
		r := *pa % *pb
		return c.FromInt64(r), xtruth.True
	case xtruth.X:
		return c.Var(), xtruth.X
	}
	return c.Zero(), xtruth.X
}

func (c consts) Band(a, b I) I {
	pa, pb := a.(*C).p, b.(*C).p
	if pa == nil || pb == nil {
		return c.Var()
	}
	z := *pa & *pb
	return c.FromInt64(z)
}

func (c consts) Bnot(a I) I {
	pa := a.(*C).p
	if pa == nil {
		return c.Var()
	}
	z := ^(*pa)
	return c.FromInt64(z)
}

func (c consts) Lshift(a, s I) (I, xtruth.T) {
	panic("unimplemented")
}

func (c consts) Rshift(a, s I) (I, xtruth.T) {
	panic("unimplemented")
}

func (c consts) Less(a, b I) xtruth.T {
	pa, pb := a.(*C).p, b.(*C).p
	if pa == nil || pb == nil {
		return xtruth.X
	}
	if *pa < *pb {
		return xtruth.True
	}
	return xtruth.False
}

func (c consts) Equal(a, b I) xtruth.T {
	pa, pb := a.(*C).p, b.(*C).p
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
