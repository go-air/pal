# Modeling golang.org/x/tools/go/ssa with pal

## Field

returns a field, no derefernce involved.

## FieldAddr(p pointer, i int), p a pointer to a struct

return the address of the i'th field of the struct
pointed to by p.


## Index

array, index

- [x] const indices
- [x] non-const indices

## IndexAddr(o point, x Expr), p a pointer to an (array or slice?).

- [ ] slice
- [ ] array
	- [x] const indices
	- [ ] non-const indices
	

## Functions

Each function object is associated with 

1. a reference memory.Loc, indicating "the" location of the function
1. a types.Signature in T.
1. if it is a method then a receiver node of appropriate type,
   then an associated memory.Loc
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

We apply calls to 'defers' and 'go f()'s and plain calls.

In all cases, the problem is looking up the function to call.  This 
needs to take into account 

- lookup of declared functions
- dynamic functions (functions which are assigned to a variable)
- method invocation.

The lookup will result in a set of possible functions to call.  The
pointer analysis, being inclusion based, will just call all of them.

## 







