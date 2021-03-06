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
	declName string
	free     []memory.Loc
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
	var err error
	h := hdr{&f.object}
	err = h.PlainEncode(w)
	if err != nil {
		return err
	}
	err = plain.Put(w, " ")
	if err != nil {
		return err
	}
	if f.declName != "" {
		_, err = fmt.Fprintf(w, "%s ", f.declName)
		if err != nil {
			return err
		}
	} else {
		_, err = w.Write([]byte("- "))
		if err != nil {
			return err
		}
	}
	err = plain.Put(w, " ")
	if err != nil {
		return err
	}

	err = plain.Uint(len(f.free)).PlainEncode(w)
	if err != nil {
		return err

	}
	for _, free := range f.free {
		err = plain.Put(w, " ")
		if err != nil {
			return err
		}
		err = free.PlainEncode(w)
		if err != nil {
			return err
		}
	}
	err = plain.Put(w, " ")
	if err != nil {
		return err
	}
	err = plain.EncodeJoin(w, " ", f.recv, plain.Uint(len(f.params)))
	if err != nil {
		return err
	}
	if f.variadic {
		err = plain.Put(w, "*")
	} else {
		err = plain.Put(w, "-")
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
	_, err := fmt.Fscanf(r, "%s ", &f.declName)
	if err != nil {
		return err
	}
	if f.declName == "-" {
		f.declName = ""
	}
	n := plain.Uint(0)
	err = (&n).PlainDecode(r)
	if err != nil {
		return err
	}
	f.free = make([]memory.Loc, n)
	for i := range f.free {
		err = plain.Expect(r, " ")
		if err != nil {
			return err
		}
		fp := &f.free[i]
		err = fp.PlainDecode(r)
		if err != nil {
			return err
		}
	}
	err = plain.Expect(r, " ")
	if err != nil {
		return err
	}
	err = plain.DecodeJoin(r, " ", &f.recv, &n)
	if err != nil {
		return fmt.Errorf("func decode join [obj recv n]: %w", err)
	}
	var buf [1]byte
	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		return err
	}
	switch buf[0] {
	case '*':
		f.variadic = true
	case '-':
		f.variadic = false
	default:
		return fmt.Errorf("unexpected modifier '%c'", buf[0])
	}
	f.params = make([]memory.Loc, n)
	for i := plain.Uint(0); i < n; i++ {
		err = plain.Expect(r, " ")
		if err != nil {
			return fmt.Errorf("func decode param %d: %w\n", i, err)
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
			return fmt.Errorf("func decode result %d: %w", i, err)
		}
	}
	return nil
}
