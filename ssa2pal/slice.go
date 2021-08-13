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

package ssa2pal

import (
	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/memory"
	"github.com/go-air/pal/results"
	"golang.org/x/tools/go/ssa"
)

type Slice struct {
	object
	Len   ssa.Value
	Cap   ssa.Value
	Slots []Slot
}

func newSlice(bld *results.Builder, l, c ssa.Value) *Slice {
	return nil
}

type Slot struct {
	I   indexing.I
	Loc memory.Loc
}
