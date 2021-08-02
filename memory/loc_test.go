package memory

import (
	"testing"

	"github.com/go-air/pal/internal/plain"
)

func TestLoc(t *testing.T) {
	org := Loc(10011)
	m := org
	p := &m
	if err := plain.EncodeDecode(p); err != nil {
		t.Fatal(err)
	}
	if *p != org {
		t.Fatalf("%s != %s\n", plain.String(p), plain.String(org))
	}
}

func TestLittleLoc(t *testing.T) {
	org := loc{parent: Loc(10011), class: Heap, attrs: IsOpaque}
	m := org
	p := &m
	if err := plain.EncodeDecode(p); err != nil {
		t.Fatal(err)
	}
	if p.parent != org.parent || p.class != org.class || p.attrs != org.attrs {
		t.Fatalf("%s != %s\n", plain.String(p), plain.String(&org))
	}
}
