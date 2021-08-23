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

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/memory"
)

func TestStruct(t *testing.T) {
	u := &Struct{}
	u.loc = 11
	u.typ = 14
	u.fields = make([]memory.Loc, 2)
	u.fields[0] = 555
	u.fields[1] = 2
	var err error
	err = plain.TestRoundTrip(u, true)
	if err != nil {
		t.Error(err)
	}

}
