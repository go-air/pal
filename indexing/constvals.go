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

	"github.com/go-air/pal/internal/byteorder"
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

func (c consts) Var(v I) bool {
	return false
}

func (c consts) FromInt(v int) I {
	return v
}

func (c consts) AsInt(v I) (int, bool) {
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

func (c consts) Less(a, b I) AbsTruth {
	aa, bb := a.(int), b.(int)
	if aa < bb {
		return True
	}
	return False
}

func (c consts) Equal(a, b I) AbsTruth {
	aa, bb := a.(int), b.(int)
	if aa == bb {
		return True
	}
	return False
}

func (c consts) PalEncodeValue(w io.Writer, v I) error {
	buf := make([]byte, 8)
	bo := byteorder.ByteOrder()
	bo.PutUint64(buf, uint64(v.(int)))
	_, e := w.Write(buf)
	return e
}

func (c consts) PalDecodeValue(r io.Reader) (I, error) {
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	bo := byteorder.ByteOrder()
	if n != 8 {
		return nil, fmt.Errorf("PalDecodeValue: couldn't read 8")
	}
	return int(bo.Uint64(buf)), nil
}

func (c consts) PalDecode(r io.Reader) error {
	return nil
}

func (c consts) PalEncode(w io.Writer) error {
	return nil
}
