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

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/memory"
)

type Struct struct {
	object
	fields []memory.Loc
}

func (s *Struct) NumFields(i int) int {
	return len(s.fields)
}

func (s *Struct) Field(i int) memory.Loc {
	return s.fields[i]
}

func (s *Struct) PlainEncode(w io.Writer) error {
	var err error
	err = plain.Put(w, "r")
	if err != nil {
		return err
	}
	err = s.object.PlainEncode(w)
	if err != nil {
		return err
	}
	for i := range s.fields {
		err = plain.Put(w, " ")
		if err != nil {
			return err
		}
		f := &s.fields[i]
		err = f.PlainEncode(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Struct) PlainDecode(r io.Reader) error {
	var err error
	err = plain.Expect(r, "r")
	if err != nil {
		return err
	}
	obj := &s.object
	err = obj.PlainDecode(r)
	if err != nil {
		return err
	}
	for i := range s.fields {
		err = plain.Expect(r, " ")
		if err != nil {
			return err
		}
		f := &s.fields[i]
		err = f.PlainDecode(r)
		if err != nil {
			return err
		}
	}
	return nil
}
