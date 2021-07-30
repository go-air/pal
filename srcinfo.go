package pal

import (
	"go/token"
)

type SrcKind int

const (
	TypeVar SrcKind = iota
	Func
	MakeArray
	MakeSlice
	MakeChan
	MakeInterface
	MakeClosure
	New
	AddressOf
	Param // do we need this with TypeVar?
)

type SrcInfo struct {
	Kind     SrcKind
	Pos      token.Pos
	FuncName string
	Param    int
}
