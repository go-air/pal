# pal functional architecture


## Memory Models

Memory is modelled as a set of nodes (as in nodes in a graph), called _locs_.
For example, locs may correspond to global variables, local variables, the
result of calls to 'malloc', function declarations, etc.  Locs may also be
specific to control flow and/or call flow context.  However, pal leaves this
opaque to the user, at least at the level of the memory model.

Sets of locs may or may not support non-constant values for their size.  For non-constant
values which occur in the program under analysis, a special Value type is provided and
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
all the values in the map are abstracted to a single pair.

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

## Values

Traditional pointer analyses such as Anderson, SteensGaard are independent of
numerical analysis.  Often such analysis are useful because they can bootstrap a
numerical analysis and they are usually much faster (albeit less precise) than
methods which combine numerical and points-to analysis.

pal provides opaque support for numerical constraints in a _Values_ type, defined
below.

```go
type Values interface {
	ToInt(v Value) (int, bool)
	FromInt(int) V
	Plus(a, b Value) Value
	Less(a, b Value) AbsTrue
	Equal(a, b Value) AbsTruth
}
```


The idea is that the pointer operations only use addition and tests for 
Values in order to implement the Mems interface; however, Values in programs
may be arbitrary expressions in the target language, which, over the 
set of all possible executions of the program, may contain any sort of
concrete value.  

A client of pal must decide how to model these concrete values, however any such
model will provide the Values interface above.

pal will provide some basic models

### Const Values

Constant values, corresponding to types\' offsets.  In this model, every load or store
to a Mem with non-constant offset are collapsed onto a single Mem with zero offset.  
This is an abstraction which is simple, efficient, and imprecise for containers
containing lots of pointers.  

Note however that the choice of value abstraction effects how the Mems are
generated.

### Intervals

TODO(wsc0)

## Solving

Suppose we have a program or a fragment of a program for which we have created
Mems, Constraints, and Values.  We would like to compute the points to set of
Mems  

In pal, all these scenarios share a common _Solver_ interface specified below.


```go

// Construct a solver from Mems (and so with the associated constraints)
// and a modelling of the values.  Results are precomputed.
func SolverForAll(ms Mems, vs Values) Solver
// Results are on demand.
func LazySolver(ms Mems, vs Values) Solver
// Results are pre-ordered according to 'perm'
func OrderedSolver(ms Mems, []int perm, vs Values) Solver

// Results are selected from q and PointsTo means transitively to
// things related to q (forward and backward )
func SelectFwdSolver(ms Mems, q []Mem, vs Values) Solver {...}
func SelectBwdSolver(ms Mems, q []Mem, vs Values) Solver {...}
func SelectSolve(ms Mems, q []Mem, vs Values) Solver {...}

// project the transitive closue of the points to onto 'on'
func ProjectedSolver(ms Mems, on []Mem, vs Values) Solver

type Solver interface {

    // Overlaps determines complex aliasing.
	Overlaps(m Mem, mext Value, n Mem, next Value) AbsTruth

	// m == n ?
	Equal(m, n Mem) AbsTruth


	// PointsTo place the points to set of m into dst, starting
	// at offset from with a max of 'ext',
	//
	// return the resulting dst.
	PointsTo(dst []Mem, m Mem, ext Value) []Mem

	// ReplaceOpaque
	// for every Mem in the underlying Mems whose points-to set
	// includes the points to set of 't', remove the points-to 
	// set of 't' and add the PointsTo set of every rep in 'reps'
	ReplaceOpaque(t Mem, reps ...Mem)

}
```
