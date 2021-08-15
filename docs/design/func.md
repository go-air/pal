# pal functional architecture


## Memory Models

Memory is modelled as a set of nodes (as in nodes in a graph), called _locs_.
For example, locs may correspond to global variables, local variables, the
result of calls to 'malloc', function declarations, etc.  Locs may also be
specific to control flow and/or call flow context.  However, pal leaves this
opaque to the user, at least at the level of the memory model.

Sets of locs may or may not support non-constant values for their size.  For non-constant
index which occur in the program under analysis, a special Value type is provided and
detailed below.

Sets of locs must provide an efficient means to determine if two locs 'm', 'n' may overlap
and whether they are equal.

The application context will generate a set (or sets) of locs, corresponding to
the source code which it analysies.

### Role in Applications

The relation between locs and target language under analysis is unspecified in pal,
in order to allow more flexible usage.  However, here we give some possible use cases.

In a classial Anderson style analysis, locs correspond to program variables
and/or locations in the program of heap allocations such as `malloc`.   

In Go's x/tools/go/pointer, a loc correponds to a node, so for example a 
map would would be represented by a node representing a (key, value) pair;
all the index in the map are abstracted to a single pair.

For flow sensitivity, an SSA form in combination with such a representation
can be used.  For more flow sensitivity an SSI form can be used.  These
representations generate variants of variables for different places in the
control flow.  By associating locs with such variants of variables, we
get some degree of flow sensitivity.


## Constraints: Refs, Loads, Stores, Transfer

```go
type Constraints interface {
	Mems() Mems
	Ref(p, v Mem) // a = &b => b in {|a|} 
	Load(dst, src Mem) // dst = *src => for all t in {|src|}, {|t|} \subseteq {|dst|}
	Store(dst, src Mem) // *dst = src => for all t in {|dst|}, {|src|} \subseteq {|t|}
	Transfer(dst, src Mem) // dst = src (without dereference) => {|src|} \subseteq {|dst|}
}
```

## index

Traditional pointer analyses such as Anderson, SteensGaard are independent of
numerical analysis.  Often such analysis are useful because they can bootstrap a
numerical analysis and they are usually much faster (albeit less precise) than
methods which combine numerical and points-to analysis.

pal provides opaque support for numerical constraints in a _index_ type, defined
below.

```go
type index interface {
	ToInt(v Value) (int, bool)
	FromInt(int) V
	Plus(a, b Value) Value
	Less(a, b Value) AbsTrue
	Equal(a, b Value) AbsTruth
}
```


The idea is that the pointer operations only use addition and tests for 
index in order to implement the Mems interface; however, index in programs
may be arbitrary expressions in the target language, which, over the 
set of all possible executions of the program, may contain any sort of
concrete value.  

A client of pal must decide how to model these concrete index, however any such
model will provide the index interface above.

pal will provide some basic models

### Const index

Constant index, corresponding to types\' offsets.  In this model, every load or store
to a Mem with non-constant offset are collapsed onto a single Mem with zero offset.  
This is an abstraction which is simple, efficient, and imprecise for containers
containing lots of pointers.  

Note however that the choice of value abstraction effects how the Mems are
generated.

### Intervals

TODO(wsc0)

## Modularity

Pal follows Go's packages, which form an acyclic dependency graph.  Pal
results are stored on a per-package basis.


### Coding

This consists of encoding the package into the pal memory model.
Notably, for modularity, we need to keep track of exportable
symbols, opaqueness, param and return index of functions.

### Solving

This consists of finding the points-to relation for the package
in terms of how it was coded.

### Export

This consists of projecting the memory model onto i/o relations
between exported opaque memory locations (including parameters
and returns of exported functions).

### Import


### Example

Below is an example diamond 
shaped package dependency graph which we will use to describe how
modular solving works.

```
S -> A
S -> B
A -> D
B -> D
```
Below is a list of actions for solving the points-to 
incrementally


1. D: Code 
1. D: Solve
1. D: Export
1. A: Import D
1. A: Code
1. A: Solve
1. A: Export
1. B: Import D
1. B: Code
1. B: Solve
1. B: Export
1. S: Import A
1. S: Import B
1. S: Code
1. S: Solve




## Solving



