package cosem

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

type DateTime time.Time

func decodeDateTime(buf *bytes.Buffer) (Data, error) {
	bytes := make([]byte, 12)
	n, err := buf.Read(bytes)
	if n != 12 || err != nil {
		return nil, fmt.Errorf("Failed to read date-time bytes")
	}

	y := int(binary.BigEndian.Uint16(bytes[0:2]))
	m := time.Month(bytes[2])
	d := int(bytes[3])

	hour := int(bytes[5])
	min := int(bytes[6])
	sec := int(bytes[7])
	nsec := 0
	if bytes[8] != 0xff {
		nsec = int(bytes[8]) * 10 * 1000 * 1000
	}

	loc := time.Local
	if dev := binary.BigEndian.Uint16(bytes[9:11]); dev != 0x8000 {
		offset := int16(dev) * 60
		loc = time.FixedZone("", int(offset))
	}

	return DateTime(time.Date(y, m, d, hour, min, sec, nsec, loc)), nil
}

func (dt DateTime) String() string {
	return time.Time(dt).String()
}
