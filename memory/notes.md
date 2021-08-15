# memory model

## structured data

struct S { f Tf, g Tg, } -> os sd(Tf) sd(Tg), os.vsz == sd(Tf).vsz + sd(Tg).vsz + 1

Arrays: non const indices are possible in the source.

We can add an accumulator node and use it.

## slices

var o type T[]
makeslice T len cap

p -> o
o contains 
	- len sink node  o.L with vsize
	- cap sink node  o.C with vsize
	- nodes for structured data for T rooted at o.D
	- o.S root node whose el
	

## 
