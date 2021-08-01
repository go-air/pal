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

package mem

type Attrs byte

const (
	Opaque Attrs = 1 << iota
	IsParam
	IsFunc
	IsReturn
)

func (a Attrs) IsOpaque() bool {
	return a&Opaque != 0
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
		byte('p'),
		boolByte(a.IsParam()),
		byte('f'),
		boolByte(a.IsFunc()),
		byte('r'),
		boolByte(a.IsReturn())})
}
