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

// Package memory defines the memory model of pal and associated operations.
//
// The memory model is basically a directed graph, where the nodes represent
// disjoint memory regions (such as the locations of local variables, or the
// code location of an allocation), and the edges represents a "may points to"
// relation.  In other words for nodes a,b where (a,b) is an edge, the memory
// model indicates that in some execution of the analysed program, a may point
// to b.
//
// Separate memory models are intended to be associated with every package
// under analysis (and we need depended upon packages transitively as well).
//
// The above description is a simplification of what is implemented here.
// First, nodes represent structured data.  Second, the edges are coded as a
// set of constraints rather than explicitly.  Third, each model represents a
// single (possibly main) package.  Fourth, the node sizes are parameterised on
// a values type which may be used to model sizes and indices.  Finally, some
// nodes are marked "Opaque" indicating that they represent unknown pointers.
// Correspondingly, there is a (work in progress) mechanism to compose models
// accross packages which allows eliminating local variables and substituting
// opaque nodes with results in a calling package.
package memory
