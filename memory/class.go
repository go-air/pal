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
)

// Class is a memory class.  Each location has a unique memory class.
type Class byte

const (
	// Zero is the nil, the only pointer value which cannot be dereferenced.
	Zero Class = iota
	// Global is the location of a global variable
	Global
	// Local is the location of a local variable
	Local
	// Heap is the location associated with a heap allocation
	Heap
	// Rel is a relative location.  Rel has parent=root and
	// non constant vsz, indicating the offset w.r.t. the root.
	//
	// wsc: this will be necessary for unsafe.Pointer conversion
	// and pointer arithmetic.
	//
	// Also, we can use it to model slice index addresses &slice[x].
	//Rel
)

func (c Class) String() string {
	switch c {
	case Zero:
		return "z"
	case Local:
		return "l"
	case Global:
		return "g"
	case Heap:
		return "h"
	default:
		panic("bad MemClass")
	}
}

func (c Class) PlainEncode(w io.Writer) error {
	_, e := w.Write([]byte(c.String()))
	return e
}

func (c *Class) PlainDecode(r io.Reader) error {
	buf := make([]byte, 1)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return err
	}
	switch buf[0] {
	case byte('z'):
		*c = Zero
	case byte('l'):
		*c = Local
	case byte('g'):
		*c = Global
	case byte('h'):
		*c = Heap
	default:
		return fmt.Errorf("pal/memory.Class: unexpected class code: %s", string(buf))
	}
	return nil
}
