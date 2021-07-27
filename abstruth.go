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

type AbsTruth int

const (
	False AbsTruth = iota
	True
	Unknown
)

func (t AbsTruth) String() string {
	switch t {
	case False:
		return "f"
	case True:
		return "t"
	case Unknown:
		return "x"
	default:
		panic("bad abstruth")
	}
}

func (t AbsTruth) Or(o AbsTruth) AbsTruth {
	if t == True || o == True {
		return True
	}
	if t == Unknown || o == Unknown {
		return Unknown
	}
	return False
}

func (t AbsTruth) And(o AbsTruth) AbsTruth {
	if t == False || o == False {
		return False
	}
	if t == Unknown || o == Unknown {
		return Unknown
	}
	return True
}

func (t AbsTruth) Not() AbsTruth {
	switch t {
	case True:
		return False
	case False:
		return True
	case Unknown:
		return Unknown
	default:
		panic("bad BasTruth")
	}
}
