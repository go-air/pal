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
	"math"

	"github.com/go-air/pal/internal/plain"
	"github.com/go-air/pal/memory"
	"github.com/go-air/pal/values"
)

// ForPkg represents results for a package.
type ForPkg struct {
	PkgPath  string
	Values   values.T
	Start    memory.Loc
	MemModel *memory.Model // provides memory.Loc operations
	SrcInfo  []SrcInfo     // indexed by memory.Loc
}

func NewForPkg(pkgPath string, vs values.T) *ForPkg {
	mdl := memory.NewModel(vs)
	return &ForPkg{
		PkgPath:  pkgPath,
		Values:   vs,
		Start:    memory.Loc(1),
		MemModel: mdl,
		SrcInfo:  make([]SrcInfo, mdl.Len())}
}

func (pkg *ForPkg) set(m memory.Loc, info *SrcInfo) {
	n := memory.Loc(uint32(cap(pkg.SrcInfo)))
	if m < n {
		pkg.SrcInfo[m] = *info
		return
	}
	if m > math.MaxUint32/2 {
		n = math.MaxUint32
	} else {
		for n <= m {
			n *= 2
		}
	}
	infos := make([]SrcInfo, n)
	copy(infos, pkg.SrcInfo)
	infos[m] = *info
	pkg.SrcInfo = infos
}

func (pkg *ForPkg) PlainEncode(w io.Writer) error {
	if _, e := fmt.Fprintf(w, "%s:%s:%d\n", pkg.PkgPath, plain.String(pkg.Start), len(pkg.SrcInfo)); e != nil {
		return e
	}
	N := pkg.MemModel.Len()
	if N > len(pkg.SrcInfo) {
		return fmt.Errorf("corrupted pkg info, length mismatch %d %d\n", N, len(pkg.SrcInfo))
	}
	for i := 0; i < N; i++ {
		si := &pkg.SrcInfo[i]
		codr := pkg.MemModel.PlainCoderAt(i)
		if _, e := fmt.Fprintf(w, "%s %s\n", plain.String(codr), plain.String(si)); e != nil {
			return nil
		}
	}
	return nil
}

func (pkg *ForPkg) PlainDecode(r io.Reader) error {
	br := bufio.NewReader(r)
	var err error
	pkg.PkgPath, err = br.ReadString(':')
	if err != nil {
		return err
	}
	var start int
	var n int
	_, err = fmt.Scanf("%d:%d\n", &start, &n)
	if err != nil {
		return err
	}
	pkg.Start = memory.Loc(start)
	pkg.MemModel.Cap(n)
	pkg.SrcInfo = make([]SrcInfo, n)
	spaceBuf := make([]byte, 1)
	for i := 0; i < n; i++ {
		loc := pkg.MemModel.PlainCoderAt(i)
		si := &pkg.SrcInfo[i]
		if err = loc.PlainDecode(r); err != nil {
			return err
		}
		_, err = io.ReadFull(r, spaceBuf)
		if err != nil {
			return err
		}
		if err = si.PlainDecode(r); err != nil {
			return err
		}
		_, err = io.ReadFull(r, spaceBuf)
		if err != nil {
			return err
		}
	}
	return nil
}
