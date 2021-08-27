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
	"testing"

	"github.com/go-air/pal/memory"
)

func TestStruct(t *testing.T) {
	u := newStruct(11, 16)
	u.fields = make([]memory.Loc, 2)
	u.fields[0] = 555
	u.fields[1] = 2
	var err error
	err = testRoundTrip(u, func(o Object) {

		u := o.(*Struct)
		u.loc = 0
		u.typ = 0
		u.fields = nil
	}, false)
	if err != nil {
		t.Error(err)
	}

}
