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

package indexing

import (
	"github.com/go-air/pal/xtruth"
)

type consts struct{}

func ConstVals() T {
	return consts{}
}

func (c consts) Zero() I {
	return I(0)
}

func (c consts) One() I {
	return I(1)
}

func (c consts) Var(v I) bool {
	return false
}

func (c consts) FromInt(v int) I {
	return v
}

func (c consts) AsInt(v I) (int, bool) {
	vv, ok := v.(int)
	if !ok {
		return 0, false
	}
	return vv, true
}

func (c consts) Plus(a, b I) I {
	aa, bb := a.(int), b.(int)
	return I(aa + bb)
}

func (c consts) Less(a, b I) xtruth.T {
	aa, bb := a.(int), b.(int)
	if aa < bb {
		return xtruth.True
	}
	return xtruth.False
}

func (c consts) Equal(a, b I) xtruth.T {
	aa, bb := a.(int), b.(int)
	if aa == bb {
		return xtruth.True
	}
	return xtruth.False
}
