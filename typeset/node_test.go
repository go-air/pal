package typeset

import (
	"testing"

	"github.com/go-air/pal/internal/plain"
)

func TestNamed(t *testing.T) {
	n := &named{}
	n.name = "f"
	n.typ = Int64
	n.loff = 0
	plain.TestRoundTrip(n, false)

}
