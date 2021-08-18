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

type node struct {
	kind     Kind
	lsize    int     // memory model logical size
	elem     Type    // pointer, array, slice
	key      Type    // map keys, method receivers
	fields   []named // struct/interface only
	params   []named // name == "" ok
	results  []named // name == "" ok
	variadic bool

	// hashing
	next Type
	hash uint32
}

// for fields, parameters, named results
type named struct {
	name string
	typ  Type
	loff int // 0 when for params or methods
}

func (n named) PlainEncode(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%s %08x", n.name, n.typ)
	return err
}

func (n *named) PlainDecode(r io.Reader) error {
	m, err := fmt.Fscanf(r, "%s %08x", &n.name, &n.typ)
	if err != nil {
		return fmt.Errorf("decode named n=%d: %w", m, err)
	}
	return nil
}

func (n *node) PlainEncode(w io.Writer) error {
	var err error
	if err = n.kind.PlainEncode(w); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, " %08x ", n.lsize)
	if err != nil {
		return err
	}
	switch n.kind {
	case Basic:
		panic("basic types are hard coded")
	case Pointer, Slice, Chan, Array:
		_, err = fmt.Fprintf(w, "%08x", n.elem)
	case Map:
		_, err = fmt.Fprintf(w, "%08x %08x", n.key, n.elem)
	case Struct, Interface:
		err = wrapJoinEncode(w, "{", ", ", "}", n.fields)

	case Func:
		var variadicString = "-"
		if n.variadic {
			variadicString = "+"
		}
		if n.key != NoType {
			_, err = fmt.Fprintf(w, "m%s%08x.", variadicString, n.key)
		} else {
			_, err = fmt.Fprintf(w, "f%s", variadicString)
		}
		if err != nil {
			return err
		}
		err = wrapJoinEncode(w, "(", ", ", ") ", n.params)
		if err != nil {
			return err
		}
		err = wrapJoinEncode(w, "(", ", ", ")", n.results)
	case Tuple:
		err = wrapJoinEncode(w, "(", ", ", ")", n.fields)
	}
	return err
}

func (n *node) PlainDecode(r io.Reader) error {
	var err error
	kp := &n.kind
	if err = kp.PlainDecode(r); err != nil {
		return err
	}
	lsz := &n.lsize
	if _, err = fmt.Fscanf(r, " %08x ", lsz); err != nil {
		return err
	}
	switch n.kind {
	case Basic:
		panic("basic types are hard coded")
	case Pointer, Slice, Chan, Array:
		_, err = fmt.Fscanf(r, "%08x", &n.elem)
	case Map:
		_, err = fmt.Fscanf(r, "%08x %08x", &n.key, &n.elem)
	case Struct, Interface:
		err = wrapJoinDecode(r, "{", ", ", "}", &n.fields)
	case Tuple:
		err = wrapJoinDecode(r, "(", ", ", ")", &n.fields)
	case Func:
		err = n.decodeFunc(r)
	}
	return err
}

func (n *node) decodeFunc(r io.Reader) error {
	buf := make([]byte, 2)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return fmt.Errorf("fn hdr: %w", err)
	}
	n.key = NoType
	switch buf[0] {
	case 'f':
	case 'm':
		_, err = fmt.Fscanf(r, "%08x", &n.key)
		if err != nil {
			return fmt.Errorf("key %w", err)
		}
	default:
		return fmt.Errorf("unexpected '%s'", string(buf))
	}
	switch buf[1] {
	case '+':
		n.variadic = true
	case '-':
		n.variadic = false
	default:
		return fmt.Errorf("unexpected '%s'", string(buf))
	}
	err = wrapJoinDecode(r, "(", ", ", ") ", &n.params)
	if err != nil {
		return fmt.Errorf("params %w", err)
	}
	err = wrapJoinDecode(r, "(", ", ", ")", &n.results)
	if err != nil {
		return fmt.Errorf("results %w", err)
	}

	return nil
}

func wrapJoinEncode(w io.Writer, left, sep, right string, elts []named) error {
	_, err := fmt.Fprintf(w, left)
	if err != nil {
		return err
	}
	sbuf := []byte(sep)
	N := len(elts)
	for i := range elts {
		err = elts[i].PlainEncode(w)
		if err != nil {
			return err
		}

		if i < N-1 {
			_, err = w.Write(sbuf)
			if err != nil {
				return err
			}
		}
	}
	_, err = fmt.Fprintf(w, right)
	return err
}

type bb struct {
	r   io.Reader
	b   []byte // size 1
	buf bool
}

func newbb(r io.Reader) *bb {
	return &bb{r: r, b: make([]byte, 1), buf: false}
}

func (bb *bb) Read(d []byte) (int, error) {
	n := 0
	var nn int
	var err error
	if bb.buf && len(d) > 0 {
		d[0] = bb.b[0]
		bb.buf = false
		d = d[1:]
		n++
		nn, err = bb.r.Read(d)
	} else {
		nn, err = bb.r.Read(d)
	}
	return nn + n, err
}

func (bb *bb) PeekByte() (byte, error) {
	if bb.buf {
		return 0, fmt.Errorf("already read %s", string(bb.b))
	}

	_, err := io.ReadFull(bb.r, bb.b)
	if err != nil {
		return 0, err
	}
	bb.buf = true
	return bb.b[0], nil
}

func wrapJoinDecode(r io.Reader, left, sep, right string, elts *[]named) error {
	br := newbb(r)
	err := expect(r, left)
	if err != nil {
		return err
	}
	for {
		b, err := br.PeekByte()
		if err != nil {
			return fmt.Errorf("peek %w", err)
		}
		switch b {
		case sep[0]:
			err = expect(br, sep)
		case right[0]:
			return expect(br, right)
		default:
			n := len(*elts)
			*elts = append(*elts, named{})
			elt := &(*elts)[n]
			err = elt.PlainDecode(br)
		}
		if err != nil {
			return fmt.Errorf("sep or dec: %w\n", err)
		}
	}
}

func expect(r io.Reader, what string) error {
	buf := []byte(what)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return err
	}
	if string(buf) != what {
		return fmt.Errorf("unexpected '%s' want '%s'", string(buf), what)
	}
	return nil
}

func (n *node) zero() {
	n.kind = Basic
	n.elem = NoType
	n.key = NoType
	n.lsize = 1
	n.fields = nil
	n.params = nil
	n.results = nil
	n.variadic = false
	n.hash = 0
	n.next = NoType
}

var basicNodes = []node{
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // NoType
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Bool
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Uint8
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Uint16
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Uint32
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Uint64
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Int8
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Int16
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Int32
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Int64
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Float32
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Float64
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Complex64
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // Complex128
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // String
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}, // UnsafePointer
	node{kind: Basic, elem: NoType, key: NoType, lsize: 1}} // Uintptr
