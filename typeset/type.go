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

package typeset

import (
	"fmt"
	"io"
)

type Type uint32

const (
	NoType Type = iota
	Bool
	Uint8
	Uint16
	Uint32
	Uint64
	Int8
	Int16
	Int32
	Int64
	Float32
	Float64
	Complex64
	Complex128
	String
	UnsafePointer
	Uintptr
	_endType
)

func (t Type) PlainEncode(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%08x", t)
	return err

}

func (t *Type) PlainDecode(r io.Reader) error {
	_, err := fmt.Fscanf(r, "%08x", t)
	return err
}
