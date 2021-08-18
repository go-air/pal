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
	return nil
}

func (slot *Slot) PlainDecode(r io.Reader) error {
	return nil
}

func (slice *Slice) PlainEncode(w io.Writer) error {
	return nil
}

func (slice *Slice) PlainDecode(r io.Reader) error {
	return nil
}
