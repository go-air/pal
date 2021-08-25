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
	"testing"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/internal/plain"
)

func clob(codr plain.Coder) {
	c := codr.(*Constraint)
	c.Dest = 0
	c.Src = 0
	c.Kind = 0
	if c.Index != nil {
		c.Index = c.Index.Gen()
	}
}

func TestConstraint(t *testing.T) {
	var d = []Constraint{
		AddressOf(11, 32),
		Load(12, 33),
		Store(13, 34),
		TransferIndex(14, 34, indexing.ConstVals().FromInt64(11))}
	for i := range d {
		if err := plain.TestRoundTripClobber(&d[i], clob, true); err != nil {
			t.Error(err)
		}
	}
}
