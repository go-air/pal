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

package pal

type MemClass uint64

const (
	Zero MemClass = iota
	Global
	Local
	Heap
	Opaque
)

func (c MemClass) String() string {
	switch c {
	case Zero:
		return "z"
	case Local:
		return "l"
	case Global:
		return "g"
	case Heap:
		return "h"
	case Opaque:
		return "?"
	default:
		panic("bad MemClass")
	}
}
