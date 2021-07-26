// Copyright 2021 Scott Cotton
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

import "fmt"

type Mem uint64

func (m Mem) Class() MemClass {
	switch m & 7 {
	case 0:
		return Zero
	case 1:
		return Oob
	case 2:
		return Thunk
	case 3:
		return Global
	case 4:
		return Local
	case 5:
		return Alloc
	case 6:
		return Ext
	default:
		panic("bad Mem")
	}
}

func (m Mem) Index() uint64 {
	return uint64(m >> 3)
}

func (m Mem) String() string {
	return fmt.Sprintf("%s%d", m.Class(), m.Index())
}
