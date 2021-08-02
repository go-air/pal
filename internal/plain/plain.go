package plain

import (
	"bytes"
	"io"

	"strings"
)

type T struct {
	Encoding
}

func (t T) Plain() string {
	var b bytes.Buffer
	if err := t.PlainEncode(&b); err != nil {
		panic(err)
	}
	return b.String()
}

func (t T) ParsePlain(s string) error {
	return t.PlainDecode(strings.NewReader(s))
}

type PlainEncoder interface {
	PlainEncode(io.Writer) error
}

type PlainDecoder interface {
	PlainDecode(io.Reader) error
}

type Encoding interface {
	PlainEncoder
	PlainDecoder
}
