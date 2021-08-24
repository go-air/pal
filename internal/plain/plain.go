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
	"strconv"

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
			_, err = w.Write(sepBytes)
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
			return fmt.Errorf("decode join elt %d: %w", i, err)
		}
	}
	return nil
}

var alpha = []byte{
	'0': 'g',
	'1': 'h',
	'2': 'i',
	'3': 'j',
	'4': 'k',
	'5': 'l',
	'6': 'm',
	'7': 'n',
	'8': 'o',
	'9': 'p',
	'a': 'q',
	'b': 'r',
	'c': 's',
	'd': 't',
	'e': 'u',
	'f': 'v',
	'g': '0',
	'h': '1',
	'i': '2',
	'j': '3',
	'k': '4',
	'l': '5',
	'm': '6',
	'n': '7',
	'o': '8',
	'p': '9',
	'q': 'a',
	'r': 'b',
	's': 'c',
	't': 'd',
	'u': 'e',
	'v': 'f'}

func EncodeInt64(w io.Writer, v int64) error {
	const N = 17 // 16 + sign
	var buf [N]byte
	x := strconv.AppendInt(buf[:0], v, 16)
	n := len(x)
	x[n-1] = alpha[x[n-1]]
	_, err := w.Write(x)
	return err
}

func DecodeInt64From(r io.Reader) (int64, error) {
	u := int64(0)
	err := DecodeInt64(r, &u)
	return u, err
}

func DecodeInt64(r io.Reader, p *int64) error {
	var buf [17]byte
	i := 0
	var err error
	var c byte
	for i < 17 {
		_, err = io.ReadFull(r, buf[i:i+1])
		c = buf[i]
		i++
		if err != nil {
			return err
		}
		if c == '-' {
			continue
		}

		if c < byte('0') {
			return fmt.Errorf("not plain int fmt")
		}
		if alpha[c] == 0 {
			return fmt.Errorf("not plain int fmt")
		}
		if c > byte('f') {
			buf[i-1] = alpha[buf[i-1]]
			break
		}
	}
	if i == 0 {
		panic("impossible")
	}
	v, e := strconv.ParseInt(string(buf[:i]), 16, 64)
	if e != nil {
		panic(e)
	}
	*p = v
	return nil
}

type Int int64

func (i Int) PlainEncode(w io.Writer) error {
	return EncodeInt64(w, int64(i))
}
func (i *Int) PlainDecode(r io.Reader) error {
	v := int64(*i)
	err := DecodeInt64(r, &v)
	if err != nil {
		return err
	}
	*i = Int(v)
	return nil
}

func EncodeUint64(w io.Writer, v uint64) error {
	const N = 16
	var buf [N]byte
	x := strconv.AppendUint(buf[:0], v, 16)
	n := len(x)
	x[n-1] = alpha[x[n-1]]
	_, err := w.Write(x)
	return err
}

func DecodeUint64From(r io.Reader) (uint64, error) {
	u := uint64(0)
	err := DecodeUint64(r, &u)
	return u, err
}

func DecodeUint64(r io.Reader, p *uint64) error {
	var buf [16]byte
	i := 0
	var err error
	var c byte
	for i < 16 {
		_, err = io.ReadFull(r, buf[i:i+1])
		if err != nil {
			return err
		}
		c = buf[i]
		i++

		if c < byte('0') {
			return fmt.Errorf("%d '%c' '%s' not plain int fmt", i, c, string(buf[:i]))
		}
		if alpha[c] == 0 {
			return fmt.Errorf("%c not plain int fmt s", c)
		}
		if c > byte('f') {
			buf[i-1] = alpha[buf[i-1]]
			break
		}
	}
	if i == 0 {
		panic("impossible")
	}
	v, _ := strconv.ParseUint(string(buf[:i]), 16, 64)
	*p = v
	return nil
}

type Uint uint64

func (u Uint) PlainEncode(w io.Writer) error {
	return EncodeUint64(w, uint64(u))
}
func (u *Uint) PlainDecode(r io.Reader) error {
	v := uint64(*u)
	err := DecodeUint64(r, &v)
	if err != nil {
		return err
	}
	*u = Uint(v)
	return nil
}

type Coder interface {
	Encoder
	Decoder
}

func TestRoundTrip(c Coder, verbose bool) error {
	return TestRoundTripClobber(c, nil, verbose)
}

func TestRoundTripClobber(c Coder, clob func(Coder), verbose bool) error {
	var buf = new(bytes.Buffer)
	if err := c.PlainEncode(buf); err != nil {
		return fmt.Errorf("round trip, enc1: %w", err)
	}
	d := buf.Bytes()
	s1 := string(d)
	if verbose {
		fmt.Printf("encoded\n```\n%s```\n", s1)
	}
	if clob != nil {
		clob(c)
	}
	buf = bytes.NewBuffer(d)
	if err := c.PlainDecode(buf); err != nil {
		return fmt.Errorf("round trip, dec: %w", err)
	}
	buf = new(bytes.Buffer)
	if err := c.PlainEncode(buf); err != nil {
		return fmt.Errorf("round trip, enc2: %w", err)
	}
	d = buf.Bytes()
	s2 := string(d)
	if verbose {
		fmt.Printf("encode2\n```\n%s```\n", s2)
	}
	if s1 != s2 {
		return fmt.Errorf("\n%s\n!=\n%s\n", s1, s2)
	}
	return nil
}

func Expect(r io.Reader, s string) error {
	buf := []byte(s)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return err
	}
	if string(buf) != s {
		return fmt.Errorf("expected '%s' got '%s'", s, buf)
	}
	return nil
}

func Put(w io.Writer, s string) error {
	_, err := w.Write([]byte(s))
	return err
}
