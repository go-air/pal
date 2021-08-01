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

package values

import (
	"fmt"
	"io"

	"github.com/go-air/pal/internal/byteorder"
)

type Const int

type consts struct{}

func ConstVals() T {
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

func (c consts) FromInt(v int) V {
	return Const(v)
}

func (c consts) AsInt(v V) (int, bool) {
	vv, ok := v.(Const)
	if !ok {
		return 0, false
	}
	return int(vv), true
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

func (c consts) PalEncodeValue(w io.Writer, v V) error {
	buf := make([]byte, 8)
	bo := byteorder.ByteOrder()
	bo.PutUint64(buf, uint64(v.(Const)))
	_, e := w.Write(buf)
	return e
}

func (c consts) PalDecodeValue(r io.Reader) (V, error) {
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	bo := byteorder.ByteOrder()
	if n != 8 {
		return nil, fmt.Errorf("PalDecodeValue: couldn't read 8")
	}
	return Const(int(bo.Uint64(buf))), nil
}

func (c consts) PalDecode(r io.Reader) error {
	return nil
}

func (c consts) PalEncode(w io.Writer) error {
	return nil
}