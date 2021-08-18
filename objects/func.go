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
	"io"

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

func (f *Func) NParams() int {
	return len(f.params)
}

func (f *Func) ResultLoc(i int) memory.Loc {
	return f.results[i]
}

func (f *Func) NResults() int {
	return len(f.results)
}

func (f *Func) PlainEncode(w io.Writer) error {
	return nil
}

func (f *Func) PlainDecode(r io.Reader) error {
	return nil
}
