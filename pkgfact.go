package pal

import "go/token"

type PkgFact struct {
	PackageName  string
	MemSourceLoc []MemSourceLoc
}

func (p *PkgFact) AFact() {}

type MemSourceLocKind int

const (
	Param MemSourceLocKind = iota
	Decl
	Alloc
)

type MemSourceLoc struct {
	Kind     MemSourceLocKind
	FuncName string
	Param    int
	Pos      token.Pos
}
