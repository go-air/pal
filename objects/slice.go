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

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/memory"
)

type Slice struct {
	object
	Len   indexing.I
	Cap   indexing.I
	slots []Slot
}

func (slice *Slice) NumSlots() int {
	return len(slice.slots)
}

func (slice *Slice) Slot(i int) Slot {
	return slice.slots[i]
}

type Slot struct {
	I   indexing.I
	Loc memory.Loc
}

func (slot *Slot) PlainEncode(w io.Writer) error {
	return plain.EncodeJoin(w, " ", slot.I, &slot.Loc)
}

func (slot *Slot) PlainDecode(r io.Reader) error {
	return plain.DecodeJoin(r, " ", slot.I, &slot.Loc)
}

func (slice *Slice) PlainEncode(w io.Writer) error {
	var err error
	err = plain.Put(w, "s")
	if err != nil {
		return err
	}
	err = slice.object.PlainEncode(w)
	if err != nil {
		return err
	}
	N := len(slice.slots)
	err = plain.EncodeUint64(w, uint64(N))
	if err != nil {
		return err
	}
	for i := 0; i < len(slice.slots); i++ {
		err = plain.Put(w, " ")
		if err != nil {
			return err
		}
		pslot := &slice.slots[i]
		err = pslot.PlainEncode(w)
		if err != nil {
			return err
		}

	}
	return nil
}

func (slice *Slice) PlainDecode(r io.Reader) error {
	var err error
	err = plain.Expect(r, "s")
	if err != nil {
		return err
	}
	pobj := &slice.object
	err = pobj.PlainDecode(r)
	if err != nil {
		return err
	}
	N := uint64(0)
	err = plain.DecodeUint64(r, &N)
	n := int(N)
	slice.slots = make([]Slot, n)
	for i := 0; i < n; i++ {
		pslot := &slice.slots[i]
		err = plain.Expect(r, " ")
		if err != nil {
			return err
		}
		err = pslot.PlainDecode(r)
		if err != nil {
			return err
		}
	}
	return nil
}
