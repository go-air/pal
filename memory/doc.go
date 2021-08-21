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
// under analysis.
//
// The above description is a simplification of what is implemented here.
// First, nodes represent structured data (arrays and structs).  Second, the
// edges are coded as a set of constraints rather than explicitly. Third, the
// node sizes are parameterised on a index type which may be used to model
// sizes and indices.
//
// Finally, some nodes are marked "Opaque" indicating that they represent
// unknown pointers.  Correspondingly, there is a (work in progress) mechanism
// to compose models accross packages which allows eliminating local variables
// and substituting opaque nodes with results in a calling package.
//
// Locs
//
// Memory locations are the fundamental unit of a memory model, defined in the
// type Loc as a uint32. Each location has a root, and root locations represent
// canonical memory regions.  A region is disjoint from another region if the
// associated roots are not identical.
//
// Internally, Locs have this structure:
//
//  type loc struct { class ...  attrs ...  root Loc; parent Loc; lsz  int; typ typeset.Type }
//
// The values of this internal structure are only accessible from the Model type.
// Memory location classes indicate whether the memory is global, local (stack), or heap
// allocated.  Memory location attributes indicate whether a location corresponds
// to a parameter, a return, and whether it is opaque.
//
// Structured Data
//
// Data structuring models structs and arrays (but not slices or maps) and how
// such data types map to contiguous memory regions.  A logical size is
// associated with each location, and is defined by the type associated with
// the data.  For struct 's' with fields '[f1, f2, ...]',  the models contain
// one location for 's': 'loc(s)',  and one location for each field 'fi':
// loc(fi).  The virtual size of 'loc(s)' is equal to 1 + the sum of the
// virtual sizes for the fields.  Array work likewise, keyed by constant index.
// Non structured types have size '1 + Sum(lsz children) == 1'
//
// Each location 'm' in structured data has a root node acting as an identifier
// for its region and a parent node, indicating to what structure it belongs.
// The parent pointers are self-loops for roots.
//
// Constraints
//
// Each model has a set of associated constraints.  Constraints have a
// collecting semantics (they act like generators or production rules in a
// grammar).
//
// PointsTo constraints 'p = &v' indicate that v is in the points to set of p.
//
// Load constraints 'd = *p' indicate that for any v in the points to set of p,
// the points-to set of v is contained in in the points to set of d,
// recursively descending structured data at *p in tandem with d.
//
// Store constraints '*p = v' indicate that for any d in the points to set of p
// and any w in the points to set of v, w is in the points to set of d,
// recursively descending structured data at v, in tandem with d.
//
// Transfer constraints 'dst = src + i' indicate the points to set of src at
// index i is contained in the points to set of dst.  i may either be a
// constant or an expression from the program under analysis.  i must be the
// constant 0 for any pointer to a basic type. i must be a constant if
// src is a pointer to a struct.  i may be an int64 expression if src is a
// pointer to an array or slice.
//
package memory
