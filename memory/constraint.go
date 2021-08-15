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

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/internal/plain"
)

type ConstraintKind int

const (
	KAddressOf ConstraintKind = iota
	KLoad
	KStore
	KTransfer
)

var ck2s = map[ConstraintKind]string{
	KAddressOf: "ad",
	KLoad:      "ld",
	KStore:     "st",
	KTransfer:  "tr"}

var s2ck = map[string]ConstraintKind{
	"ad": KAddressOf,
	"ld": KLoad,
	"st": KStore,
	"tr": KTransfer}

func (ck ConstraintKind) PlainEncode(w io.Writer) error {
	_, err := w.Write([]byte(ck2s[ck]))
	return err
}

func (ck *ConstraintKind) PlainDecode(r io.Reader) error {
	buf := make([]byte, 2)
	_, e := io.ReadFull(r, buf)
	if e != nil {
		return e
	}
	s := string(buf)
	k, present := s2ck[s]
	if !present {
		return fmt.Errorf("unknown constraint kind '%s'", s)
	}
	*ck = k
	return nil
}

type Constraint struct {
	Kind  ConstraintKind
	Dest  Loc
	Src   Loc
	Index indexing.I
}

func AddressOf(dst, src Loc) Constraint {
	return Constraint{Kind: KAddressOf, Dest: dst, Src: src}
}

func Load(dst, src Loc) Constraint {
	return Constraint{Kind: KLoad, Dest: dst, Src: src}
}

func Store(dst, src Loc) Constraint {
	return Constraint{Kind: KStore, Dest: dst, Src: src}
}

func Transfer(dst, src Loc) Constraint {
	return TransferIndex(dst, src, 0)
}

func TransferIndex(dst, src Loc, i indexing.I) Constraint {
	return Constraint{Kind: KTransfer, Dest: dst, Src: src, Index: i}
}

func (c *Constraint) PlainEncode(w io.Writer) error {
	return plain.EncodeJoin(w, " ", c.Kind, c.Dest, c.Src)
}

func (c *Constraint) PlainDecode(r io.Reader) error {
	return plain.DecodeJoin(r, " ", &c.Kind, &c.Dest, &c.Src)
}
