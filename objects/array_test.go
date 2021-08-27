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

func clobArray(o Object) {
	a := o.(*Array)
	a.loc = 0
	a.typ = 0
	a.n = 0
	a.elemSize = 0
}

func TestArray(t *testing.T) {
	a := newArray(3, 7)
	a.n = 17
	a.elemSize = 5
	if err := testRoundTrip(a, clobArray, false); err != nil {
		t.Error(err)
	}
}
