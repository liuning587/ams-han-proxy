package hdlc

import "fmt"

type Address struct {
	Logical  int
	Physical int
}

func (a *Address) UnmarshalBinary(data []byte) error {
	switch len(data) {
	case 1:
		a.Logical = int(data[0] >> 1)
		a.Physical = 0
	case 2:
		a.Logical = int(data[0] >> 1)
		a.Physical = int(data[1] >> 1)
	case 4:
		a.Logical = (int(data[0]>>1) << 7) | int(data[1]>>1)
		a.Physical = (int(data[2]>>1) << 7) | int(data[3]>>1)
	default:
		return fmt.Errorf("Invalid address length (%d)", len(data))
	}
	return nil
}

func (a *Address) String() string {
	return fmt.Sprintf("(%d,%d)", a.Physical, a.Logical)
}
