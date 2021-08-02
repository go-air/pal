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

package plain

import (
	"bytes"
	"io"

	"strings"
)

type T struct {
	Encoding
}

func (t T) Plain() string {
	var b bytes.Buffer
	if err := t.PlainEncode(&b); err != nil {
		panic(err)
	}
	return b.String()
}

func (t T) ParsePlain(s string) error {
	return t.PlainDecode(strings.NewReader(s))
}

type PlainEncoder interface {
	PlainEncode(io.Writer) error
}

type PlainDecoder interface {
	PlainDecode(io.Reader) error
}

type Encoding interface {
	PlainEncoder
	PlainDecoder
}