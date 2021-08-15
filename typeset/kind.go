package typeset

type Kind uint32

const (
	_     Kind = iota
	Basic      = iota
	Pointer
	Array
	Struct
	Slice
	Map
	Chan
	Interface
	Func
	Tuple
)
