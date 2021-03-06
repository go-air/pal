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

	"github.com/go-air/pal/memory"
	"github.com/go-air/pal/typeset"
)

type Pointer struct {
	object
}

func newPointer(loc memory.Loc, typ typeset.Type) *Pointer {
	return &Pointer{object: object{kind: kpointer, loc: loc, typ: typ}}
}

func (p *Pointer) PlainEncode(w io.Writer) error {
	return (&p.object).plainEncode(w)
}

func (p *Pointer) plainDecode(r io.Reader) error {
	return nil
}
