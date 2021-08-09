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

// Package plain provides interfaces and a few supporting functions for
// a 'plain' encoding.
//
// A 'plain' encoding should serialize data in a plain text way, without
// being too 'pretty'.
package plain

import (
	"bytes"
	"fmt"
	"io"

	"strings"
)

// String provides a String() function
// for Encoders.
func String(t Encoder) string {
	var b bytes.Buffer
	if err := t.PlainEncode(&b); err != nil {
		panic(err)
	}
	return b.String()
}

// Parse provides a Parse() function
// for decoders.
func Parse(t Decoder, s string) error {
	return t.PlainDecode(strings.NewReader(s))
}

// Encoder is the interface for a plain encoder.
type Encoder interface {
	PlainEncode(io.Writer) error
}

// Decoder is the interface for a plain decoder.
type Decoder interface {
	PlainDecode(io.Reader) error
}

func EncodeJoin(w io.Writer, sep string, es ...Encoder) error {
	sepBytes := []byte(sep)
	var err error
	for i, e := range es {
		if i != 0 {
			_, err = fmt.Fprint(w, sepBytes)
			if err != nil {
				return err
			}
		}
		err = e.PlainEncode(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func DecodeJoin(r io.Reader, sep string, ds ...Decoder) error {
	var err error
	buf := make([]byte, len(sep))
	for i, d := range ds {
		if i != 0 {
			_, err = io.ReadFull(r, buf)
			if string(buf) != sep {
				return fmt.Errorf("unexpected: '%s'", string(buf))
			}
		}
		err = d.PlainDecode(r)
		if err != nil {
			return err
		}
	}
	return nil
}

type Coder interface {
	Encoder
	Decoder
}

func EncodeDecode(c Coder) error {
	var buf = new(bytes.Buffer)
	if err := c.PlainEncode(buf); err != nil {
		return err
	}
	d := buf.Bytes()
	buf = bytes.NewBuffer(d)
	return c.PlainDecode(buf)
}
