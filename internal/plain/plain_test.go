package plain

import (
	"bytes"
	"math"
	"testing"
)

func TestInt(t *testing.T) {
	vs := []int64{math.MinInt64, -1, 0, 1, math.MaxInt64}
	for _, v := range vs {
		i := Int(v)
		TestRoundTrip(&i, false)
		buf := bytes.NewBuffer(nil)
		err := Int(v).PlainEncode(buf)
		if err != nil {
			t.Error(err)
			return
		}
		vv := Int(v)
		vvp := &vv
		buf = bytes.NewBuffer(buf.Bytes())
		err = vvp.PlainDecode(buf)
		if err != nil {
			t.Error(err)
			return
		}
		if int64(vv) != v {
			t.Errorf("mismatch int64: %d != %d\n", int64(v), vv)
		}
	}
}

func TestUint(t *testing.T) {
	vs := []uint64{0, 1, 17, 444447, math.MaxUint64}
	for _, v := range vs {
		i := Uint(v)
		TestRoundTrip(&i, false)
		buf := bytes.NewBuffer(nil)
		err := Uint(v).PlainEncode(buf)
		if err != nil {
			t.Error(err)
			return
		}
		vv := Uint(v)
		vvp := &vv
		buf = bytes.NewBuffer(buf.Bytes())
		err = vvp.PlainDecode(buf)
		if err != nil {
			t.Error(err)
			return
		}
		if uint64(vv) != v {
			t.Errorf("mismatch int64: %d != %d\n", int64(v), vv)
		}
	}
}
