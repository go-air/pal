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
	"go/types"
	"testing"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/internal/plain"
)

func clobSlice(c plain.Coder) {
	s := c.(*Slice)
	s.loc = 0
	s.typ = 0
	s.slots = nil
}

func TestSlot(t *testing.T) {
	idx := indexing.ConstVals()
	slot := &Slot{Ptr: 17, Obj: 3, I: idx.Var()}
	plain.TestRoundTripClobber(slot, func(c plain.Coder) {
		o := c.(*Slot)
		o.Ptr = 2
		o.Obj = 3
		o.I = idx.Var()
	}, true)

}

func TestSlice(t *testing.T) {

	idx := indexing.ConstVals()
	b := NewBuilder("testslice", idx)
	s := b.Slice(types.NewSlice(types.Typ[types.Int]), idx.Var(), idx.Var())
	b.AddSlot(s, idx.Var())
	//s := &Slice{}
	//s.loc = 3
	//s.typ = 7
	//var v1 indexing.I = idx.Var()
	//var v2 indexing.I = idx.Var()
	//fmt.Printf("v1 %p v2 %p\n", v1, v2)
	//s.slots = []Slot{
	//	{Ptr: 17, Obj: 3, I: v1},
	//	{Ptr: 18, Obj: 1, I: v2}}
	if err := plain.TestRoundTripClobber(s, clobSlice, true); err != nil {
		t.Error(err)
	}
}
