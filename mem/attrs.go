package mem

type Attrs byte

const (
	Opaque Attrs = 1 << iota
	IsParam
	IsFunc
	IsReturn
)

func (a Attrs) IsOpaque() bool {
	return a&Opaque != 0
}

func (a Attrs) IsParam() bool {
	return a&IsParam != 0
}

func (a Attrs) IsFunc() bool {
	return a&IsFunc != 0
}

func (a Attrs) IsReturn() bool {
	return a&IsReturn != 0
}

func boolByte(b bool) byte {
	if b {
		return byte('+')
	}
	return byte('-')
}
func (a Attrs) String() string {
	return string([]byte{
		byte('o'),
		boolByte(a.IsOpaque()),
		byte('p'),
		boolByte(a.IsParam()),
		byte('f'),
		boolByte(a.IsFunc()),
		byte('r'),
		boolByte(a.IsReturn())})
}
