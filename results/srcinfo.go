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
	"go/token"
	"io"

	"github.com/go-air/pal/internal/byteorder"
)

type SrcKind int

const (
	TypeVar SrcKind = iota
	Func
	MakeArray
	MakeSlice
	MakeChan

	MakeInterface
	MakeClosure

	New
	AddressOf
)

// SrcInfo represents the information about
// the source code pal expects to extract
// from an ir such as golang.org/x/tools/go/ssa.
type SrcInfo struct {
	Kind SrcKind
	Pos  token.Pos
}

// PalEncode encodes si onto w, returning
// a non-nil error if there was a problem
// writing.
func (si *SrcInfo) PalEncode(w io.Writer) error {
	buf := make([]byte, 9)
	buf[0] = byte(si.Kind)
	byteorder.ByteOrder().PutUint64(buf[1:], uint64(si.Pos))
	_, e := w.Write([]byte{byte(si.Kind)})
	return e
}

// PalDecode decodes r into si, overwriting
// si's fields
func (si *SrcInfo) PalDecode(r io.Reader) error {
	buf := make([]byte, 9)
	var s, n int
	var err error
	for {
		n, err = r.Read(buf)
		s += n
		if s == 9 && (err == nil || err == io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		buf = buf[n:]
	}
	si.Kind = SrcKind(buf[0])
	si.Pos = token.Pos(byteorder.ByteOrder().Uint64(buf[1:]))
	return nil
}
