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

package values

import (
	"io"
)

type ValueKind int

const (
	ConstKind ValueKind = iota
	VarKind
	PlusKind
)

type V interface {
}

type T interface {
	Zero() V
	One() V
	AsInt(v V) (i int, ok bool)
	FromInt(i int) V
	Var(v V) bool
	Plus(a, b V) V
	Equal(a, b V) AbsTruth
	Less(a, b V) AbsTruth
	Kind(v V) ValueKind
	PalEncodeValue(io.Writer, V) error
	PalDecodeValue(io.Reader) (V, error)
	PalEncode(io.Writer) error
	PalDecode(io.Reader) error
}
