package typeset

type Type uint32

const (
	NoType Type = iota
	Bool
	Uint8
	Uint16
	Uint32
	Uint64
	Int8
	Int16
	Int32
	Int64
	Float32
	Float64
	Complex64
	Complex128
	String
	UnsafePointer
	Uintptr
	_endType
)
