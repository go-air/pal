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

package results

import (
	"bytes"
	"go/types"
	"strings"
	"testing"

	"github.com/go-air/pal/index"
	"github.com/go-air/pal/memory"
)

func TestPlain(t *testing.T) {
	pkgres := NewPkgRes("test", index.ConstVals())
	bld := NewBuilder(pkgres)
	bld.Type = types.Typ[types.Int]
	bld.GenLoc()
	bld.Attrs = memory.IsParam
	bld.GenLoc()
	bld.Class = memory.Heap
	bld.GenLoc()
	buf := bytes.NewBuffer(nil)
	if err := pkgres.PlainEncode(buf); err != nil {
		t.Fatal(err)
	}
	s1 := buf.String()
	if err := pkgres.PlainDecode(strings.NewReader(s1)); err != nil {
		t.Fatal(err)
	}
	buf = bytes.NewBuffer(nil)
	if err := pkgres.PlainEncode(buf); err != nil {
		t.Fatal(err)
	}
	s2 := buf.String()
	if s1 != s2 {
		t.Errorf("pkgres plain:\n%s\n!=\n%s\n", s1, s2)
	}
}
