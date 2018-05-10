package cosem

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Integer int64

func decodeInteger(buf *bytes.Buffer, length int, signed bool) (Integer, error) {
	bytes := make([]byte, length)
	n, err := buf.Read(bytes)
	if n != length || err != nil {
		return 0, fmt.Errorf("Failed to read bytes from buffer")
	}
	switch length {
	case 1:
		return Integer(bytes[0]), nil
	case 2:
		return Integer(binary.BigEndian.Uint16(bytes)), nil
	case 4:
		return Integer(binary.BigEndian.Uint32(bytes)), nil
	default:
		return 0, fmt.Errorf("Invalid byte length (%d)", length)
	}
}

func (i Integer) String() string {
	return fmt.Sprint(int64(i))
}
