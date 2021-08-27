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
)

func clobMap(o Object) {
	m := o.(*Map)
	m.loc = 0
	m.typ = 0
	m.key = 0
	m.elem = 0
}

func TestMap(t *testing.T) {
	a := &Map{}
	a.loc = 3
	a.typ = 7
	a.key = 17
	a.elem = 5
	if err := testRoundTrip(a, clobMap, false); err != nil {
		t.Error(err)
	}
}
