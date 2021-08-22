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

// Chan object
//
// c.loc represents pointer
// c.slot represents things sent to
// and received from the channel.
type Chan struct {
	object
	slot memory.Loc
}

func (c *Chan) Recv(dst memory.Loc, mm *memory.Model) {
	mm.AddLoad(dst, c.loc)
}

func (c *Chan) Send(src memory.Loc, mm *memory.Model) {
	mm.AddStore(c.loc, src)
}

func (c *Chan) PlainEncode(w io.Writer) error {
	var err error
	err = plain.Put(w, "c")
	if err != nil {
		return err
	}
	if err = c.object.PlainEncode(w); err != nil {
		return err
	}
	return c.slot.PlainEncode(w)
}

func (c *Chan) PlainDecode(r io.Reader) error {
	var err error
	err = plain.Expect(r, "c")
	if err != nil {
		return err
	}
	p := &c.object
	if err = p.PlainDecode(r); err != nil {
		return err
	}
	ps := &c.slot
	return ps.PlainDecode(r)
}
