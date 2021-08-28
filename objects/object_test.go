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
	"bytes"
	"fmt"
)

// helper function for other tests, uses PlainDecode func
// and the object's PlainEncode method.  returns
// non nil error if the encoders return non nil or
// o.PlainEncode() != PlainDecode(o.PlainEncode).PlainEncode
func testRoundTrip(o Object, clobbr func(o Object), verbose bool) error {
	io := bytes.NewBuffer(nil)
	var err error
	err = o.PlainEncode(io)
	if err != nil {
		return err
	}
	s1 := io.String()
	if verbose {
		fmt.Printf("objects.testRoundTrip enc1: '%s'\n", s1)
	}
	io = bytes.NewBuffer(io.Bytes())
	obj, err := PlainDecodeObject(io)
	if err != nil {
		return err
	}
	io = bytes.NewBuffer(nil)
	err = obj.PlainEncode(io)
	if err != nil {
		return err
	}
	s2 := io.String()
	if s1 != s2 {
		return fmt.Errorf("enc -> %s, enc,dec,enc -> %s", s1, s2)
	}
	if verbose {
		fmt.Printf("objects.testRoundTrip enc1: '%s'\n", s2)
	}
	return nil
}
