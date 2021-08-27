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

type Func struct {
	object
	sig      typeset.Type
	declName string
	fnobj    memory.Loc
	recv     memory.Loc
	params   []memory.Loc
	results  []memory.Loc
	variadic bool
}

func newFunc(loc memory.Loc, typ typeset.Type) *Func {
	return &Func{object: object{kind: kfunc, loc: loc, typ: typ}}
}

func (f *Func) Declared() bool {
	return f.declName != ""
}

func (f *Func) Name() string {
	return f.declName
}

func (f *Func) Loc() memory.Loc {
	return f.loc
}

func (f *Func) RecvLoc(i int) memory.Loc {
	return f.recv
}

func (f *Func) ParamLoc(i int) memory.Loc {
	return f.params[i]
}

func (f *Func) NumParams() int {
	return len(f.params)
}

func (f *Func) ResultLoc(i int) memory.Loc {
	return f.results[i]
}

func (f *Func) NumResults() int {
	return len(f.results)
}

func (f *Func) PlainEncode(w io.Writer) error {
	err := plain.EncodeJoin(w, " ", hdr{&f.object},
		f.fnobj, f.recv, plain.Uint(len(f.params)))
	if f.variadic {
		err = plain.Put(w, "*")
	} else {
		err = plain.Put(w, ".")
	}
	if err != nil {
		return err
	}
	for _, p := range f.params {
		err = plain.Put(w, " ")
		if err != nil {
			return err
		}
		err = p.PlainEncode(w)
		if err != nil {
			return err
		}
	}
	// results
	err = plain.Put(w, " ")
	if err != nil {
		return err
	}
	err = plain.Uint(len(f.results)).PlainEncode(w)
	if err != nil {
		return err
	}
	for _, res := range f.results {
		err = plain.Put(w, " ")
		if err != nil {
			return err
		}
		err = res.PlainEncode(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Func) plainDecode(r io.Reader) error {
	err := plain.Expect(r, " ")
	if err != nil {
		return err
	}
	n := plain.Uint(0)
	err = plain.DecodeJoin(r, " ", &f.fnobj, &f.recv, &n)
	if err != nil {
		return err
	}
	var buf [1]byte
	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		return err
	}
	switch buf[0] {
	case '*':
		f.variadic = true
	case '.':
		f.variadic = false
	default:
		return fmt.Errorf("unexpected modifier '%c'", buf[0])
	}
	f.params = make([]memory.Loc, n)
	for i := plain.Uint(0); i < n; i++ {
		err = plain.Expect(r, " ")
		if err != nil {
			return err
		}
		p := &f.params[i]
		err = p.PlainDecode(r)
		if err != nil {
			return err
		}
	}
	err = plain.Expect(r, " ")
	if err != nil {
		return fmt.Errorf("unexpected space 2: %w\n", err)
	}
	err = (&n).PlainDecode(r)
	if err != nil {
		return err
	}

	f.results = make([]memory.Loc, n)
	for i := plain.Uint(0); i < n; i++ {
		err = plain.Expect(r, " ")
		if err != nil {
			return err
		}
		p := &f.results[i]
		err = p.PlainDecode(r)
		if err != nil {
			return err
		}
	}
	return nil
}
