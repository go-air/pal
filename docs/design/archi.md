# pal architecture

The pal architecture is centered around the idea of  _persistant modular
analysis_.

This means that the analysis is bottom-up in the dependency graph, mirroring
the Go build system and the go/tools analysis library.  However, the analysis
is _persistent_, meaning the results are stored and may be used later to
different ends.  

In this bottom-up phase, pal constructs executable and queryable memory models
for each exported symbol of each package. 

## Applications

### Command line

### Module proxy

 
## Related Work

pal is fundamentally based on a Anderson analysis [3], however it introduces a
symbolic aspect for treatment of numerics, and a _meta symbolic_ aspect for
incrementality.  pal is designed to be retargeted and adaptable to different
applications, and provides a mechanism for modular analysis.

srcPtr [7] works on the AST is a framework for Anderson or Steensgaard analysis.
pal is completely agnostic of the input: it could be AST or some IR such as
SSA or SSI.  pal does not directly require a call graph or a control flow graph,
it is lower level and only provides the pointer related operations.  Of course
we anticipate that such operations are normally called during a traversal of a 
program representation, but no assumptions are made about that representation.

Infer [2] is a bug finding tool which performs compositional memory analysis 
using _biabduction_.  Like Infer, pal (atleast in compositional usage) uses
meta-symbolic variables (variables whose values are program variables) to 
reason about functions without knowing anything about call sites.  Unlike Infer, 
pal does not use separation logic, rather the heap is modelled as a graph
between memory locations. 

Gillian [5] is also language agnostic, however it is based on modelling full
programs by symbolic execution in a given IR (GIL) whereas pal only
symbolically executes the numeric _Values_ in pointer arithmetic, allowing the
caller to model these values in many different ways.

Golang's pointer analysis [8] is Anderson style with less flexibility,
dependency on an ssa package specific to Go.

## Out of Scope

pal does not attempt to model edge case scenarios in Go, such as
correct pointer analysis in the presence of races or unsafe.Pointer usage.

Rather, pal should provide a small set of basic operations which, taken together,
can be used to model a variety of program behaviors while focusing principally
on usage for which Go guarantees memory safety. 
