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

package typeset

import (
	"fmt"
	"io"
)

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

var kind2string = map[Kind]string{
	Basic:     "bas",
	Pointer:   "ptr",
	Array:     "arr",
	Struct:    "str",
	Slice:     "sli",
	Map:       "map",
	Chan:      "chn",
	Interface: "ifa",
	Func:      "fun",
	Tuple:     "tup"}

var string2kind = map[string]Kind{
	"bas": Basic,
	"ptr": Pointer,
	"arr": Array,
	"str": Struct,
	"sli": Slice,
	"map": Map,
	"chn": Chan,
	"ifa": Interface,
	"fun": Func,
	"tup": Tuple}

func (k Kind) String() string {
	return kind2string[k]
}
func (k Kind) PlainEncode(w io.Writer) error {
	_, err := w.Write([]byte(kind2string[k]))
	return err
}

func (k *Kind) PlainDecode(r io.Reader) error {
	buf := make([]byte, 3)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return err
	}
	kk, present := string2kind[string(buf)]
	if !present {
		return fmt.Errorf("unknown kind: %s", string(buf))
	}
	*k = kk
	return nil
}
