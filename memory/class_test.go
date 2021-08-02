package memory

import (
	"testing"

	"github.com/go-air/pal/internal/plain"
)

func TestClass(t *testing.T) {
	org := Heap
	class := org
	p := &class
	if err := plain.EncodeDecode(p); err != nil {
		t.Fatal(err)
	}
	if *p != org {
		t.Fatalf("%s != %s\n", plain.String(p), plain.String(org))
	}
}
