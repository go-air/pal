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

package objects

import (
	"io"

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/memory"
)

type Array struct {
	object
	elemSize int64
	n        int64
}

func (a *Array) Len() int64 {
	return a.n
}

func (a *Array) At(i int) memory.Loc {
	z := a.loc + 1
	return z + (memory.Loc(a.elemSize) * memory.Loc(i))
}

func (a *Array) PlainEncode(w io.Writer) error {
	var err error
	err = plain.Put(w, "a")
	if err != nil {
		return err
	}
	return plain.EncodeJoin(w, " ", &a.object, plain.Uint(a.elemSize), plain.Uint(a.n))
}

func (a *Array) PlainDecode(r io.Reader) error {
	var err error
	err = plain.Expect(r, "a")
	if err != nil {
		return err
	}
	esz, n := plain.Uint(a.elemSize), plain.Uint(a.n)
	err = plain.DecodeJoin(r, " ", &a.object, &esz, &n)
	if err != nil {
		return err
	}
	a.elemSize = int64(esz)
	a.n = int64(n)
	return nil
}
