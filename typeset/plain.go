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

func (t *T) PlainEncode(w io.Writer) error {
	N := len(t.nodes)
	_, err := fmt.Fprintf(w, "%d:%d\n", N-int(_endType), cap(t.hash))
	if err != nil {
		return err
	}
	for i := int(_endType); i < N; i++ {
		node := &t.nodes[i]
		if err = node.PlainEncode(w); err != nil {
			return fmt.Errorf("typeset encode node %d: %w", i, err)
		}
		_, err = fmt.Fprintf(w, "\n")
		if err != nil {
			return fmt.Errorf("typeset encode eol %d: %w", i, err)
		}
	}
	return nil
}

func (t *T) PlainDecode(r io.Reader) error {
	tt := New()
	var N int
	var H int
	_, err := fmt.Fscanf(r, "%d:%d\n", &N, &H)
	if err != nil {
		return fmt.Errorf("typeset decode hdr: %w", err)
	}
	tt.hash = make([]Type, H)
	copy(tt.hash[:_endType], t.hash[:_endType])
	tt.nodes = make([]node, N+int(_endType), H)
	copy(tt.nodes[:_endType], t.nodes[:_endType])
	eol := []byte("\n")
	for i := 0; i < N; i++ {
		ty := Type(i) + _endType
		node := &tt.nodes[ty]
		node.zero()
		if err = node.PlainDecode(r); err != nil {
			return fmt.Errorf("typeset decode node %d: %w", ty, err)
		}
		n, err := io.ReadFull(r, eol)
		if err != nil {
			return fmt.Errorf("typeset decode eol %d: %d %w", ty, n, err)
		}
		if eol[0] != byte('\n') {
			return fmt.Errorf("expected eol got '%s'", string(eol))
		}
		node.hash = tt.hashCode(ty)
		hi := node.hash % uint32(H)
		node.next = tt.hash[hi]
		tt.hash[hi] = Type(ty)
	}
	t.nodes = tt.nodes
	t.hash = tt.hash
	return nil
}
