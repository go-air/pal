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

package memory

import (
	"fmt"
	"io"
	"strconv"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/internal/plain"
)

// Loc represents a memory location.
//
// This memory location is intented for pal points to analysis.
// it has nothing to do with the numeric value of a pointer
// in a program.
type Loc uint32

const NoLoc = Loc(0)

func (m Loc) PlainEncode(w io.Writer) error {
	_, e := fmt.Fprintf(w, "%08x", m)
	return e
}

func (m *Loc) PlainDecode(r io.Reader) error {
	var buf = make([]byte, 8)
	if _, err := io.ReadFull(r, buf); err != nil {
		return fmt.Errorf("decode loc: %w", err)
	}
	n, e := strconv.ParseUint(string(buf), 16, 32)
	*m = Loc(uint32(n))
	return e
}

type loc struct {
	class  Class
	attrs  Attrs
	root   Loc
	parent Loc

	lsz indexing.I // == 1 + Sum({c.vsz | c.parent == loc and c != loc})

	obj Loc // locals and globals are passed by addr...  NoLoc if unknown

	mark int // scratch space for internal algos
}

func (m *loc) PlainEncode(w io.Writer) error {
	return plain.EncodeJoin(w, " ", m.class, m.attrs, m.parent, m.obj)
}

func (m *loc) PlainDecode(r io.Reader) error {
	return plain.DecodeJoin(r, " ", &m.class, &m.attrs, &m.parent, &m.obj)
}
