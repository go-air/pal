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
	"fmt"
	"io"
)

type Attrs byte

const (
	IsOpaque Attrs = 1 << iota
	IsFunc
	IsParam
	IsReturn
)

const NoAttrs Attrs = 0

func (a Attrs) IsOpaque() bool {
	return a&IsOpaque != 0
}

func (a Attrs) IsParam() bool {
	return a&IsParam != 0
}

func (a Attrs) IsFunc() bool {
	return a&IsFunc != 0
}

func (a Attrs) IsReturn() bool {
	return a&IsReturn != 0
}

func (a Attrs) PlainEncode(w io.Writer) error {
	_, e := w.Write([]byte(a.String()))
	return e
}

var (
	_Attrs [4]Attrs = [4]Attrs{
		IsOpaque, IsFunc, IsParam, IsReturn}
)

func (a *Attrs) decode(buf []byte) error {
	if buf[0] != byte('o') ||
		buf[2] != byte('f') ||
		buf[4] != byte('p') ||
		buf[6] != byte('r') {
		return fmt.Errorf("expected: %s", string(buf))
	}
	*a = Attrs(0)
	for i := 1; i < 8; i += 2 {
		switch buf[i] {
		case '-':
		case '+':
			*a |= _Attrs[i/2]
		default:
			return fmt.Errorf("expected: %s", string(buf))
		}
	}
	return nil
}

func (a *Attrs) PlainDecode(r io.Reader) error {

	buf := make([]byte, 8)
	_, e := io.ReadFull(r, buf)
	if e != nil {
		return e
	}
	return a.decode(buf)
}

func boolByte(b bool) byte {
	if b {
		return byte('+')
	}
	return byte('-')
}
func (a Attrs) String() string {
	return string([]byte{
		byte('o'),
		boolByte(a.IsOpaque()),
		byte('f'),
		boolByte(a.IsFunc()),
		byte('p'),
		boolByte(a.IsParam()),
		byte('r'),
		boolByte(a.IsReturn())})
}
