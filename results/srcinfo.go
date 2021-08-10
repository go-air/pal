// Copyright 2021 The pal authors (see AUTHORS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package results

import (
	"fmt"
	"go/token"
	"io"
)

type SrcKind int

const (
	SrcVar SrcKind = iota
	SrcFunc
	SrcMakeArray
	SrcMakeSlice
	SrcMakeMap
	SrcMakeChan

	SrcMakeInterface
	SrcMakeClosure

	SrcNew
	SrcAddressOf
)

var k2s = map[SrcKind]string{
	SrcVar:           "var",
	SrcFunc:          "fun",
	SrcMakeArray:     "arr",
	SrcMakeSlice:     "sli",
	SrcMakeChan:      "chn",
	SrcMakeMap:       "map",
	SrcMakeInterface: "int",
	SrcMakeClosure:   "clo",
	SrcNew:           "new",
	SrcAddressOf:     "adr"}

var s2k = map[string]SrcKind{
	"var": SrcVar,
	"fun": SrcFunc,
	"arr": SrcMakeArray,
	"sli": SrcMakeSlice,
	"chn": SrcMakeChan,
	"map": SrcMakeMap,
	"int": SrcMakeInterface,
	"clo": SrcMakeClosure,
	"new": SrcNew,
	"adr": SrcAddressOf}

func (k SrcKind) String() string {
	return k2s[k]
}

func (k SrcKind) PlainEncode(w io.Writer) error {
	buf := []byte(k.String())
	_, e := w.Write(buf)
	return e
}

func (k *SrcKind) PlainDecode(r io.Reader) error {
	buf := make([]byte, 3)
	_, e := io.ReadFull(r, buf)
	if e != nil {
		return e
	}
	var present bool
	*k, present = s2k[string(buf)]
	if !present {
		return fmt.Errorf("unknown srckind: %s", string(buf))
	}
	return nil
}

// SrcInfo represents the information about
// the source code pal expects to extract
// from an ir such as golang.org/x/tools/go/ssa.
type SrcInfo struct {
	Kind SrcKind
	Pos  token.Pos
}

func (si *SrcInfo) PlainEncode(w io.Writer) error {
	var err error
	if err = si.Kind.PlainEncode(w); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, " %08x", si.Pos)
	return err
}

func (si *SrcInfo) PlainDecode(r io.Reader) error {
	var err error
	p := &si.Kind
	if err = p.PlainDecode(r); err != nil {
		return fmt.Errorf("si decode: %w", err)
	}
	q := &si.Pos
	_, err = fmt.Fscanf(r, " %08x", q)
	return err
}
