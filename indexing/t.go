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
	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/xtruth"
)

type I interface {
	Gen() I
	plain.Coder
}

type T[I any] interface {
	// replace with go/constant.
	Zero() I
	One() I
	// replace with go/constant?
	ToInt64(v I) (i int64, ok bool)

	FromInt64(i int64) I

	IsVar(v I) bool
	Var() I

	Plus(a, b I) I
	Times(a, b I) I

	Div(a, b I) (I, xtruth.T)
	Rem(a, b I) (I, xtruth.T)

	Band(a, b I) I
	Bnot(a I) I

	Lshift(a, s I) (I, xtruth.T)
	Rshift(a, s I) (I, xtruth.T)

	Equal(a, b I) xtruth.T
	Less(a, b I) xtruth.T

	plain.Coder
}
