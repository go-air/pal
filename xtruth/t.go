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

package xtruth

type T int

const (
	False T = iota
	True
	X
)

func (t T) String() string {
	switch t {
	case False:
		return "f"
	case True:
		return "t"
	case X:
		return "x"
	default:
		panic("bad xtruth")
	}
}

func (t T) Or(o T) T {
	if t == True || o == True {
		return True
	}
	if t == X || o == X {
		return X
	}
	return False
}

func (t T) And(o T) T {
	if t == False || o == False {
		return False
	}
	if t == X || o == X {
		return X
	}
	return True
}

func (t T) Not() T {
	switch t {
	case True:
		return False
	case False:
		return True
	case X:
		return X
	default:
		panic("bad xtruth")
	}
}
