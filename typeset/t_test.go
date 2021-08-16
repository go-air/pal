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

package typeset

import (
	"bytes"
	"go/types"
	"testing"
)

func TestTypeSet(t *testing.T) {
	ts := New()
	sli := types.NewSlice(types.Typ[types.Int])
	arr := types.NewArray(types.Typ[types.Float32], 44)
	ptr := types.NewPointer(arr)
	_ = ts.FromGoType(sli)
	_ = ts.FromGoType(arr)
	_ = ts.FromGoType(ptr)
	var buf = bytes.NewBuffer(nil)
	if err := ts.PlainEncode(buf); err != nil {
		t.Fatal(err)
	}
	str := string(buf.Bytes())
	buf = bytes.NewBuffer(buf.Bytes())
	if err := ts.PlainDecode(buf); err != nil {
		t.Fatal(err)
	}
	buf = bytes.NewBuffer(nil)
	if err := ts.PlainEncode(buf); err != nil {
		t.Fatal(err)
	}
	if str != string(buf.Bytes()) {
		t.Fatalf("\n%s\n!=\n%s\n", str, string(buf.Bytes()))
	}
}
