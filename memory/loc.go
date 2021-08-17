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
	"go/token"
	"io"
	"strconv"

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/typeset"
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
	pos    token.Pos
	root   Loc
	parent Loc
	typ    typeset.Type

	lsz int

	obj Loc // for locals and globals passed by addr...  NoLoc if unknown

	mark int // scratch space for internal algos
}

type PlainPos token.Pos

func (p PlainPos) PlainEncode(w io.Writer) error {
	_, err := fmt.Fprintf(w, "@%08x", p)
	return err
}
func (p *PlainPos) PlainDecode(r io.Reader) error {
	_, err := fmt.Fscanf(r, "@%08x", p)
	return err

}

func (m *loc) PlainEncode(w io.Writer) error {
	return plain.EncodeJoin(w, " ", m.class, m.attrs, PlainPos(m.pos), m.parent, m.obj)
}

func (m *loc) PlainDecode(r io.Reader) error {
	pp := PlainPos(m.pos)
	err := plain.DecodeJoin(r, " ", &m.class, &m.attrs, &pp, &m.parent, &m.obj)
	m.pos = token.Pos(pp)
	return err
}
