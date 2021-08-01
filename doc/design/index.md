# pal -- Pointer Analysis Library for Go

## Goals

The goal of pal is to provide a library which can be effectively used for
different kinds of pointer analyses for Go on different intermediate
representations.

### Effective pointer analysis

Pointer analysis (PA) is a core dependency of many static analyses, which have
different needs, such as

1. Providing a sound dynamic call graph. 
This in turn has many applications
	- impact analysis
 	- non-interference analysis
	- almost any interprocedural sound analysis
	- resolving method calls (with more precision)
	- dataflow analysis, eg for security
1. Identifying possible invalid pointer dereferences.
1. Proving that nil pointer dereferences or
out of bounds panics are impossible.
1. Linking traditional numeric, memory-unaware, analysis methods to modern use. 
1. Identifying aliases.

Unfortunately, PA is often or usually done under global program analysis, as
opposed to modularly.  Tools such as Golang's pointer analysis often requires
re-analyzing the standard library.  Larger projects such as Docker or
Kubernetes take even more resources.

In this project, effective pointer analysis means providing a relatively simple
api to meet the most common needs well, and to meet most needs reasonably.

### For different Go IRs

staticcheck [6] has an IR, golang.org/x/tools/go/ssa is a baseline, we are
working on (air)[https://github.com/go-air/air].  We would like pal to be
retargetable to these different IRs.  Perhaps it can be used one day for the Go
gc compiler IR, or other IRs.

However, to be standard, we will start with a golang.org/x/tools/go/ssa
implementation.

## Architecture

Please see [archi](archi.md).

## Functionality

Please see [func](func.md).



Below is the interface.



#### Role in Applications

The relation between mems and target language under analysis is unspecified in pal,
in order to allow more flexible usage.  However, here we give some possible use cases.

In a classial Anderson style analysis, Mems correspond to program variables
and/or locations in the program of heap allocations such as `malloc`.   

In Go's x/tools/go/pointer, a mem correponds to a node, so for example a 
map would would be represented by a node representing a (key, value) pair;
all the values in the map are abstracted to a single pair.

For flow sensitivity, an SSA form in combination with such a representation
can be used.  For more flow sensitivity an SSI form can be used.

There are various mechanisms which increase the precision of an analysis by
increasing the number of nodes.  Go's maps, or allocation points can have
many nodes, provided that the distinction between them is preserved in the
analysis.

One commonly pursued feature is context sensitivity, where, since a single program
point may be reached by several execution paths, the context in which it is reached
is represented  and used to create more nodes, one for each context. 
This can be done using counters, or traces, or in the most extreme case by 
parameterizing each node on the state of the heap.

pal leaves such modelling to the client application.  This is a tradeoff: on
the one hand, _any_ such modelling can use pal, so it is more flexible.  On
the other, some expertise and understanding is necessary to use pal in
any practical context.


### Refs, Loads, Stores, Transfer

For Mem 'a', 'b' in a Mem 's', we can generate a set of constraints.

```go
type Constraints interface {
	Mems() Mems
	Ref(p, v Mem) // a = &b
	Load(dst, src Mem) // dst = *src
	Store(dst, src Mem) // *dst = src
	Transfer(dst, src Mem) // dst = src (without dereference)
}
```

### Values

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

#### All Const

All maps and slices have constant size 1.


#### Const Values

Constant values, corresponding to types\' offsets.  In this model, every load or store
to a Mem with non-constant offset are collapsed onto a single Mem with zero offset.  
This is an abstraction which is simple, efficient, and imprecise for containers
containing lots of pointers.  

Note however that the choice of value abstraction effects how the Mems are
generated.

#### Linear Values [modulo]

Linearly modelled values (possibly modulo a constant to model field accesses in
arrays) can model a great deal of pointer accesses precisely.  However, the cost
of the solving necessarily goes up substantially.  Linearly modelled values may
create a lot of nodes.  This can be mitigated by having max sizes per array
value and evaluating the Value offsets modulo the fixed size (a different 
phenomenon than field access)

#### Full Target Language Expressivity

At the most precise end of the spectrum, one could imagine using symbolic execution to 
express all target language operators.  


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

	// ReplaceThunk
	// for every Mem in the underlying Mems whose points-to set
	// includes the points to set of 't', remove the points-to 
	// set of 't' and add the PointsTo set of every rep in 'reps'
	ReplaceThunk(t Mem, reps ...Mem)

}
```

## Modular/Incremental Analysis

Modular analysis, for example on a per function basis can applied as follows.
Take a sub program which contains all of its dependencies.  For any function
which is exposed, associate a Mems.Thunk() with each parameter which may contain
pointers (imagine casts!) or which is address taken.  The analysis will then be
solved in terms of the thunks:  

Once that part of the program has been analyzed, one may call ProjectedSolving
on nodes local to the calling code after a call to ReplaceThunk for each
exposed parameter.

TBD(wsc) work on this

## References

[1] Pointer Analysis. Foundations and Trends in Programming Languages Vol. 2, No. 1 (2015) 1–69
2015 Y. Smaragdakis and G. Balatsouras
DOI: 10.1561/2500000014 (https://yanniss.github.io/points-to-tutorial15.pdf)

[2] Infer
Compositional Analysis by means of bi-abduction
Journal of the ACM Volume 58 Issue 6
December 2011 
Article No.: 26pp 1–66https://doi.org/10.1145/2049697.2049700

[3] Andersen, Lars Ole (1994). Program Analysis and Specialization for the C
Programming Language (PDF) (PhD thesis).

[4] Steensgaard, Bjarne (1996). "Points-to analysis in almost linear time" (PDF). POPL '96: Proceedings of the 23rd ACM SIGPLAN-SIGACT symposium on Principles of programming languages. New York, NY, USA: ACM. pp. 32–41. doi:10.1145/237721.237727. ISBN 0-89791-769-3.

[5] @misc{maksimović2021gillian,
      title={Gillian: A Multi-Language Platform for Unified Symbolic Analysis}, 
      author={Petar Maksimović and José Fragoso Santos and Sacha-Élie Ayoun and Philippa Gardner},
      year={2021},
      eprint={2105.14769},
      archivePrefix={arXiv},
      primaryClass={cs.PL}
}

[6] [static check](staticcheck.io) 

[7] Zyrianov, Vlas; Newman, Christian D.; Guarnera, Drew T.; Collard, Michael L.; Maletic, Jonathan I. (2019). "srcPtr: A Framework for Implementing Static Pointer Analysis Approaches" (PDF). ICPC '19: Proceedings of the 27th IEEE International Conference on Program Comprehension. Montreal, Canada: IEEE.

[8] golang.org/x/tools/go/pointer
