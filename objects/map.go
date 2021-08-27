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
	"github.com/go-air/pal/typeset"
)

type Map struct {
	object
	key  memory.Loc
	elem memory.Loc
}

func newMap(loc memory.Loc, typ typeset.Type) *Map {
	return &Map{object: object{kind: kmap, loc: loc, typ: typ}}
}

func (m *Map) Key() memory.Loc {
	return m.key
}

func (m *Map) Elem() memory.Loc {
	return m.elem
}

func (m *Map) Update(k, v memory.Loc, mm *memory.Model) {
	mm.AddTransfer(m.key, k)
	mm.AddTransfer(m.elem, v)
}

func (m *Map) Lookup(dst memory.Loc, mm *memory.Model) {
	// no key transfer, it is equality check under the hood
	mm.AddTransfer(dst, m.elem)
}

func (m *Map) PlainEncode(w io.Writer) error {
	return plain.EncodeJoin(w, " ", hdr{&m.object}, m.key, m.elem)
}

func (m *Map) plainDecode(r io.Reader) error {
	var err error
	err = plain.Expect(r, " ")
	if err != nil {
		return err
	}
	return plain.DecodeJoin(r, " ", &m.key, &m.elem)
}
