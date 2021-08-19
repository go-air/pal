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

	"github.com/go-air/pal/internal/plain"
)

func TestLoc(t *testing.T) {
	org := Loc(10011)
	m := org
	p := &m
	if err := plain.TestRoundTrip(p, false); err != nil {
		t.Fatal(err)
	}
	if *p != org {
		t.Fatalf("%s != %s\n", plain.String(p), plain.String(org))
	}
}

func TestLittleLoc(t *testing.T) {
	org := loc{parent: Loc(10011), class: Heap, attrs: IsOpaque}
	m := org
	p := &m
	if err := plain.TestRoundTrip(p, false); err != nil {
		t.Fatal(err)
	}
	if p.parent != org.parent || p.class != org.class || p.attrs != org.attrs {
		t.Fatalf("%s != %s\n", plain.String(p), plain.String(&org))
	}
}
