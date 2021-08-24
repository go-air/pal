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

package indexing

import (
	"testing"

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/xtruth"
)

func TestConstVals(t *testing.T) {
	idx := ConstVals()
	z := idx.Zero()
	o := idx.One()
	v := idx.FromInt64(32)
	w := idx.Var()
	var err error
	for _, i := range [...]I{z, o, v, w} {
		err = plain.TestRoundTripClobber(i, func(c plain.Coder) {
			cc := c.(*C)
			cc.p = nil
		}, true)
		if err != nil {
			t.Error(err)
		}
	}
	switch x := idx.Equal(z, idx.Zero()); x {
	case xtruth.True:
	default:
		t.Errorf("%s == %s => %s\n", z, idx.Zero(), x)
	}
	switch x := idx.Equal(o, idx.One()); x {
	case xtruth.True:
	default:
		t.Errorf("%s == %s => %s\n", o, idx.One(), x)
	}
	switch x := idx.Equal(v, idx.FromInt64(32)); x {
	case xtruth.True:
	default:
		t.Errorf("%s == %s => %s\n", v, idx.FromInt64(32), x)
	}
}
