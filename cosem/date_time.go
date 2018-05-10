package cosem

import (
	"encoding/binary"
	"fmt"
	"time"
)

func ParseDateTime(data []byte) (time.Time, error) {
	if len(data) != 12 {
		return time.Time{}, fmt.Errorf("Invalid date-time length (%d)", len(data))
	}
	y := int(binary.BigEndian.Uint16(data[0:2]))
	m := time.Month(data[2])
	d := int(data[3])

	hour := int(data[5])
	min := int(data[6])
	sec := int(data[7])
	nsec := 0
	if data[8] != 0xff {
		nsec = int(data[8]) * 10 * 1000 * 1000
	}

	loc := time.Local
	if dev := binary.BigEndian.Uint16(data[9:11]); dev != 0x8000 {
		offset := int16(dev) * 60
		loc = time.FixedZone("", int(offset))
	}

	return time.Date(y, m, d, hour, min, sec, nsec, loc), nil
}
