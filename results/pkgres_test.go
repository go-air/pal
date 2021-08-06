package results

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
	"testing"

	"github.com/go-air/pal/memory"
	"github.com/go-air/pal/values"
)

func TestPlain(t *testing.T) {
	pkgres := NewPkgRes("test", values.ConstVals())
	bld := NewBuilder(pkgres)
	bld.Type = types.Typ[types.Int]
	bld.GenLoc()
	bld.Attrs = memory.IsParam
	bld.GenLoc()
	bld.Class = memory.Heap
	bld.GenLoc()
	buf := bytes.NewBuffer(nil)
	fmt.Printf("encode\n")
	if err := pkgres.PlainEncode(buf); err != nil {
		t.Fatal(err)
	}
	s1 := buf.String()
	t.Logf("encoded as\n\n%s\n", s1)
	fmt.Printf("decode\n")
	if err := pkgres.PlainDecode(strings.NewReader(s1)); err != nil {
		t.Fatal(err)
	}
	buf = bytes.NewBuffer(nil)
	fmt.Printf("recode\n")
	if err := pkgres.PlainEncode(buf); err != nil {
		t.Fatal(err)
	}
	s2 := buf.String()
	if s1 != s2 {
		t.Errorf("pkgres plain:\n%s\n!=\n%s\n", s1, s2)
	}
}
