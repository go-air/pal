package memory

import (
	"testing"

	"github.com/go-air/pal/internal/plain"
)

func TestAttrs(t *testing.T) {
	org := IsOpaque | IsReturn
	attrs := org
	p := &attrs
	if err := plain.EncodeDecode(p); err != nil {
		t.Fatal(err)
	}
	if *p != org {
		t.Fatalf("%s != %s\n", plain.String(p), plain.String(org))
	}
}
