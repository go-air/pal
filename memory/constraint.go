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

type Constraint[Index plain.Coder] struct {
	Kind  ConstraintKind
	Dest  Loc
	Src   Loc
	Index Index
}

func AddressOf[Index plain.Coder](dst, src Loc) Constraint[Index] {
	return Constraint[Index]{Kind: KAddressOf, Dest: dst, Src: src}
}

func Load[Index plain.Coder](dst, src Loc) Constraint[Index] {
	return Constraint[Index]{Kind: KLoad, Dest: dst, Src: src}
}

func Store[Index plain.Coder](dst, src Loc) Constraint[Index] {
	return Constraint[Index]{Kind: KStore, Dest: dst, Src: src}
}

func TransferIndex[Index plain.Coder](dst, src Loc, i Index) Constraint[Index] {
	return Constraint[Index]{Kind: KTransfer, Dest: dst, Src: src, Index: i}
}

func (c *Constraint[Index]) PlainEncode(w io.Writer) error {
	switch c.Kind {
	case KTransfer:
		return plain.EncodeJoin(w, " ", c.Kind, c.Dest, c.Src, c.Index)
	default:
		return plain.EncodeJoin(w, " ", c.Kind, c.Dest, c.Src)
	}
}

func (c *Constraint[Index]) PlainDecode(r io.Reader) error {
	err := plain.DecodeJoin(r, " ", &c.Kind, &c.Dest, &c.Src)
	if err != nil {
		return err
	}
	if c.Kind == KTransfer {
		err = plain.Expect(r, " ")
		if err != nil {
			return err
		}
		err = c.Index.PlainDecode(r)
	}
	return err
}
