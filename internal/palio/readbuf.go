package palio

import "io"

// ReadBuf reads 'r' into 'buf'.
//
// Unlike io.Reader.Read, ReadBuf will always read
// as much as possible from 'r'.  Moreover, if ReadBuf
// reads len(buf) bytes but r.Read returns io.EOF
// on the last io.Reader.Read call, ReadBuf returns
// a nil error.
//
// if n, e = ReadBuf(b, r), then n == len(b) <=> e != nil.
//
func ReadBuf(buf []byte, r io.Reader) (int, error) {
	ttl := 0
	var n int
	var e error
	var N int = len(buf)
	for ttl < N {
		n, e = r.Read(buf[ttl:])
		ttl += n
		if ttl < N {
			if e == nil {
				continue
			}
			return ttl, e
		}
		break
	}
	// don't make the caller handle EOF if we read everything.
	if e == io.EOF {
		e = nil
	}
	return ttl, e
}
