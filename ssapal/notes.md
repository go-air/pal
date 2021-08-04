# Modeling golang.org/x/tools/go/ssa with pal

## Calls

Each function object is associated with 

1. a reference node, indicating "the" location of the function
1. a types.Signature in T.
1. if it is a method then a receiver node of appropriate type,
   then an associated node
1. if it has params, then one root node per parameter
1. if it has returns, then one root node per return
1. whether or not declared
1. (if not) Free/Captured Vars (if so these are globals)


The nodes referenced above are contiguous and in the order referenced above.

## Opaqueness

If the function is package local, then no nodes are opaque.

If the function is exported, then its parameters and returns are opaque.
To enforce this, we can
1. For declared funcs, use their name
1. For dynamic funcs, 
	- if they are global, then we need to make the params/returns opaque.
	- if they are local, then if they are returned to the caller it will
	pass through an opaque node, so it should be ok to leave them not opaque.

This will be used in the import/export mechanism of memory models.

## Calls

Basic case:

args, rets






