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
	"bufio"
	"fmt"
	"io"

	"github.com/go-air/pal/indexing"
	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/memory"
)

// PkgRes represents results for a package.
type PkgRes struct {
	PkgPath  string
	indexing indexing.T
	Start    memory.Loc
	MemModel *memory.Model // provides memory.Loc operations
}

func NewPkgRes(pkgPath string, vs indexing.T) *PkgRes {
	mdl := memory.NewModel(vs)
	return &PkgRes{
		PkgPath:  pkgPath,
		indexing: vs,
		Start:    memory.Loc(1),
		MemModel: mdl}
}

func (pkg *PkgRes) PlainEncode(w io.Writer) error {
	if _, e := fmt.Fprintf(w, "%s:%s:%d\n", pkg.PkgPath, plain.String(pkg.Start), pkg.MemModel.Len()); e != nil {
		return e
	}
	N := pkg.MemModel.Len()
	for i := 0; i < N; i++ {
		codr := pkg.MemModel.PlainCoderAt(i)
		if _, e := fmt.Fprintf(w, "%s\n", plain.String(codr)); e != nil {
			return nil
		}
	}
	return pkg.MemModel.PlainEncodeConstraints(w)
}

func (pkg *PkgRes) PlainDecode(r io.Reader) error {
	br := bufio.NewReader(r)
	var err error
	pkg.PkgPath, err = br.ReadString(':')
	if err != nil {
		return fmt.Errorf("1 %w", err)
	}
	pkg.PkgPath = pkg.PkgPath[:len(pkg.PkgPath)-1]
	var start int
	var n int
	_, err = fmt.Fscanf(br, "%d:%d\n", &start, &n)
	if err != nil {
		return fmt.Errorf("2 %w", err)
	}
	pkg.Start = memory.Loc(start)
	pkg.MemModel.Cap(n)
	spaceBuf := make([]byte, 1)
	for i := 0; i < n; i++ {
		loc := pkg.MemModel.PlainCoderAt(i)
		if err = loc.PlainDecode(br); err != nil {
			return fmt.Errorf("3 %d-%w", i, err)
		}
		_, err = io.ReadFull(br, spaceBuf)
		if err != nil || spaceBuf[0] != byte('\n') {
			return fmt.Errorf("6 %d-'%s'-%w", i, string(spaceBuf), err)
		}
	}
	return pkg.MemModel.PlainDecodeConstraints(br)
}
