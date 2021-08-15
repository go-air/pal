# pal

pal -- pointer analysis library for Go

See [this blog post](https://go-air.github.io/blog/20210729-pal.html)
for an overview.

## status: volatile prototype

pal is in a volatile prototyping stage.  We have re-organised 
things a few times already, and there are many sizeable holes
still to be coded.

## roadmap

- cli
	- [x] analyzer framework stub
	- [ ] service
- [ ] memory model
	- [x] constraints (load, store, transfer)
	- [x] plain serialize
- [ ] indexing
	- [ ] integrate types
	- [ ] dev.typeparams version
- [ ] types -- represent locatable objects
	- [ ] to go types
	- [ ] from go types
	- [ ] serialize
- [ ] objects -- manage lifecycle of memory model w.r.t. Go things
	- [ ] creation
	- 
- [ ] ssa2pal
	- [x] loads
	- [x] stores
	- [x] map values to memory locations
	- [ ] map values to objects
	- [ ] indexing arithmetic operations
	- [x] structs
	- [x] arrays
	- [ ] slices
	- [x] returns
	- [x] phi nodes
	- [x] function objects
	- [ ] function variadics
	- [ ] builtins
- [ ] docs
	- [x] statement of purpose
	- [ ] design
		- [ ] cli
		- [ ] service
	- [ ] tutorial
	- [ ] reference






