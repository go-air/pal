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

package objects

import (
	"fmt"
	"io"

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/memory"
	"github.com/go-air/pal/typeset"
)

type Object interface {
	plain.Encoder
	Loc() memory.Loc
	Type() typeset.Type
}

type kind int

const (
	karray kind = iota
	kstruct
	kslice
	kmap
	kpointer
	kchan
	kinterface
	ktuple
	kfunc
)

var k2str = map[kind]string{
	karray:     "a",
	kstruct:    "u",
	kslice:     "s",
	kmap:       "m",
	kpointer:   "p",
	kchan:      "c",
	kinterface: "i",
	ktuple:     "t",
	kfunc:      "f"}

var str2k = map[string]kind{
	"a": karray,
	"u": kstruct,
	"s": kslice,
	"m": kmap,
	"p": kpointer,
	"c": kchan,
	"i": kinterface,
	"t": ktuple,
	"f": kfunc}

type object struct {
	kind kind
	loc  memory.Loc
	typ  typeset.Type
}
type hdr struct{ *object }

func (h hdr) PlainEncode(w io.Writer) error {
	return h.plainEncode(w)
}

func (k kind) PlainEncode(w io.Writer) error {
	_, err := w.Write([]byte(k2str[k]))
	return err
}

func (k *kind) PlainDecode(r io.Reader) error {
	var buf [1]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return err
	}
	kk, ok := str2k[string(buf[:])]
	if !ok {
		return fmt.Errorf("unexpected kind: '%s'", string(buf[:]))
	}
	*k = kk
	return nil
}

func (o *object) Loc() memory.Loc    { return o.loc }
func (o *object) Type() typeset.Type { return o.typ }

func (o *object) plainEncode(w io.Writer) error {
	return plain.EncodeJoin(w, " ", o.kind, o.loc, o.typ)
}

func (o *object) plainDecode(r io.Reader) error {
	return plain.DecodeJoin(r, " ", &o.kind, &o.loc, &o.typ)
}

func PlainDecode(r io.Reader) (Object, error) {
	obj := &object{}
	err := obj.plainDecode(r)
	if err != nil {
		return nil, err
	}
	fmt.Printf("decoding kind %s\n", plain.String(obj.kind))
	switch obj.kind {
	case karray:
		arr := &Array{object: *obj}
		err := arr.plainDecode(r)
		if err != nil {
			return nil, err
		}
		return arr, nil
	case kstruct:
		strukt := &Struct{object: *obj}
		err := strukt.plainDecode(r)
		if err != nil {
			return nil, err
		}
		return strukt, nil
	case ktuple:
		tuple := &Tuple{object: *obj}
		err := tuple.plainDecode(r)
		if err != nil {
			return nil, err
		}
		return tuple, nil
	case kslice:
		slice := &Slice{object: *obj}
		err := slice.plainDecode(r)
		if err != nil {
			return nil, err
		}
		return slice, nil
	case kmap:
		m := &Map{object: *obj}
		err := m.plainDecode(r)
		if err != nil {
			return nil, err
		}
		return m, nil
	case kpointer:
		ptr := &Pointer{object: *obj}
		err := ptr.plainDecode(r)
		if err != nil {
			return nil, err
		}
		return ptr, nil
	case kchan:
		chn := &Chan{object: *obj}
		err := chn.plainDecode(r)
		if err != nil {
			return nil, err
		}
		return chn, nil
	case kinterface:
		inter := &Interface{object: *obj}
		err := inter.plainDecode(r)
		if err != nil {
			return nil, err
		}
		return inter, nil
	case kfunc:
		fn := &Func{object: *obj}
		err := fn.plainDecode(r)
		if err != nil {
			return nil, err
		}
		return fn, nil
	default:
		panic("bad object kind")
	}
}
