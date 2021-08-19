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

// Package objects coordinates lower level memory and typeset
// models with Go objects.
//
// Builder
//
// The Builder is responsible for implementing the majority of
// this coordination.  It has a few different sets of methods.
//
// The first set supports a way to configure the parameters for
// the lower level memory representation.  These methods set and
// return the modified builder.  They are
//
//  1. Builder.Pos(go/token.Pos)
//  2. Builder.Type(go/types.Type)
//  3. Builder.Attrs(pal/memory.Attrs)
//  4. Builder.Class(pal/memory.Class)
//
// Making a memory location containing objects of type `T`
// representing code at pos `pos` with attributes `memory.IsOpaque`
// could for example be done like this:
//
//  b.Type(T).Pos(pos).Attrs(memory.IsOpaque).Gen()
//
// The second set of methods relates to creating locations, objects
// and constraints.  These are
//
//  1. Gen() generate a memory location.
//  2. Add{Load,Store,Transfer,PointsTo}(dst, src)
//
// The last set of methods gives access to the build results
//
//  1. Builder.Memory() *memory.Model
//  2. Builder.TypeSet() *typeset.TypeSet
//
package objects
