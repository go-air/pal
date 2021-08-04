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
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/values"
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
		return err
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

	obj Loc
	hdl Loc

	vsz values.V // == 1 + Sum({c.vsz | c.parent == loc and c != loc})

	// constraints
	pointsTo  []Loc // this loc points to that
	transfers []Loc //
	loads     []Loc // this loc = *(that loc)
	stores    []Loc // *(this loc) = that loc

	// points-to (and from)
	in  []Loc
	out []Loc
}

func (m *loc) PlainEncode(w io.Writer) error {
	_, e := fmt.Fprintf(w, "%s %s %s %s %s",
		plain.String(m.class), plain.String(m.attrs),
		plain.String(m.parent), plain.String(m.obj), plain.String(m.hdl))
	return e
}

func (m *loc) PlainDecode(r io.Reader) error {
	var word string
	var err error
	br := bufio.NewReader(r)
	word, err = br.ReadString(' ')
	if err != nil {
		return err
	}
	if err = plain.Parse(&m.class, word); err != nil {
		return err
	}
	word, err = br.ReadString(' ')
	if err := plain.Parse(&m.attrs, word); err != nil {
		return err
	}
	word, err = br.ReadString(' ')
	if err := plain.Parse(&m.parent, word); err != nil {
		return err
	}
	word, err = br.ReadString(' ')
	if err := plain.Parse(&m.obj, word); err != nil {
		return err
	}
	word, err = br.ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}
	return plain.Parse(&m.hdl, word)
}
