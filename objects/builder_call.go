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

package objects

import (
	"fmt"

	"github.com/go-air/pal/memory"
)

// dst is the memory loc of the result.
// - if f returns a single value, then that
// - otherwise it is a tuple.
//
// args indicate the arguments at the call site.
func (b *Builder) Call(f *Func, dst memory.Loc, args []memory.Loc) {
	fmt.Printf("n args %d n params %d\n", len(args), len(f.params))
	start := 0
	if f.recv != memory.NoLoc {
		start = 1
		b.AddStore(f.recv, args[0])
	}
	for i, arg := range args[start:] {
		b.AddStore(f.params[i], arg)
	}
	if dst == memory.NoLoc {
		return
	}
	switch len(f.results) {
	case 0:
	case 1:
		b.AddLoad(dst, f.results[0])
	default:
		dstTuple := b.omap[dst].(*Tuple)
		for i, ret := range f.results {
			b.AddLoad(dstTuple.At(i), ret)
		}
	}
}
