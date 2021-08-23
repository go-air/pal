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
	"fmt"
	"io"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/memory"
)

// Slices are modelled as follows.
//
// Each slice `s` has a principal pointer `ptr(s)` stored in object.loc, and a Len and a Cap
// which are of type indexing.I, and a single memory.Loc "elem" representing
// what is stored in the slice.
//
// Each slice has 0 or more slots.  A slot is a triple (p, m, i) such that
//  1. p = &m
//  2. p = ptr(s) + i
//
// For the moment, we set ptr(s) == &nil to guarantee that nil derefs
// are considered possible.  When indexing gets richer, perhaps we
// can do more.
type Slice struct {
	object
	Len   indexing.I
	Cap   indexing.I
	slots []Slot
}

func (slice *Slice) Ptr() memory.Loc {
	return slice.loc
}

func (slice *Slice) NumSlots() int {
	return len(slice.slots)
}

func (slice *Slice) Slot(i int) Slot {
	return slice.slots[i]
}

type Slot struct {
	I   indexing.I
	Ptr memory.Loc
	Obj memory.Loc
}

func (slot *Slot) PlainEncode(w io.Writer) error {
	return plain.EncodeJoin(w, " ", slot.I, &slot.Ptr, &slot.Obj)
}

func (slot *Slot) PlainDecode(r io.Reader) error {
	fmt.Printf("slot decode %p %p %p %p\n", slot, slot.I, &slot.Ptr, &slot.Obj)
	return plain.DecodeJoin(r, " ", slot.I, &slot.Ptr, &slot.Obj)
}

func (slice *Slice) PlainEncode(w io.Writer) error {
	var err error
	err = plain.Put(w, "s ")
	if err != nil {
		return err
	}
	err = slice.object.PlainEncode(w)
	if err != nil {
		return err
	}

	err = plain.Put(w, " ")
	if err != nil {
		return err
	}
	N := len(slice.slots)
	err = plain.EncodeUint64(w, uint64(N))
	if err != nil {
		return err
	}
	for i := 0; i < N; i++ {
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
	err = plain.Expect(r, "s ")
	if err != nil {
		return err
	}
	pobj := &slice.object
	err = pobj.PlainDecode(r)
	if err != nil {
		return err
	}
	err = plain.Expect(r, " ")
	if err != nil {
		return err
	}
	N := uint64(0)
	err = plain.DecodeUint64(r, &N)
	slice.slots = make([]Slot, N)
	for i := uint64(0); i < N; i++ {
		err = plain.Expect(r, " ")
		if err != nil {
			return err
		}
		pslot := &slice.slots[i]
		// XXX
		pslot.I = indexing.ConstVals().Var()
		err = pslot.PlainDecode(r)
		if err != nil {
			return err
		}
	}
	return nil
}
