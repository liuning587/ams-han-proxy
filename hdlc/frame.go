package hdlc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/howeyc/crc16"
)

// FrameDelimiter marks the start and the end of each frame
const FrameDelimiter byte = 0x7e

// SplitFrames tokenizes HDLC frames. To be used with bufio.Scanner.
func SplitFrames(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) < 3 {
		return
	}
	if data[0] != FrameDelimiter || data[1] == FrameDelimiter {
		advance = 1
		return
	}

	// Invariant: data[0] == FrameDelimiter && data[1] != FrameDelimiter
	length := int(data[1]&0x7)<<8 | int(data[2])
	if len(data) < length+1 {
		return
	}

	advance = length + 1
	token = make([]byte, length)
	copy(token, data[1:advance])
	return
}

// Frame represents a HDLC frame
type Frame struct {
	formatType byte
	length     int
	dest       []byte
	src        []byte
	ctrl       byte
	info       []byte
}

// UnmarshalBinary decodes a HDLC frame from tokenized byte arrays.
func (f *Frame) UnmarshalBinary(data []byte) error {
	n := len(data)
	buf := bytes.NewBuffer(data)

	err := f.getFormat(buf)
	if err != nil {
		return fmt.Errorf("Frame format invalid: %s", err)
	}
	if n != f.length {
		return fmt.Errorf("Byte array length (%d) doesn't match length field (%d)", n, f.length)
	}
	f.dest, err = getAddress(buf)
	if err != nil {
		return fmt.Errorf("Destination address invalid: %s", err)
	}
	f.src, err = getAddress(buf)
	if err != nil {
		return fmt.Errorf("Source address invalid: %s", err)
	}
	f.ctrl, err = buf.ReadByte()
	if err != nil {
		return fmt.Errorf("Control field invalid: %s", err)
	}
	pos := n - buf.Len()
	err = verifyChecksum(buf, crc16.ChecksumCCITT(data[0:pos]))
	if err != nil {
		return fmt.Errorf("Header checksum invalid: %s", err)
	}
	if buf.Len() > 0 {
		f.info = buf.Next(buf.Len() - 2)
		err = verifyChecksum(buf, crc16.ChecksumCCITT(data[0:n-2]))
		if err != nil {
			return fmt.Errorf("Checksum invalid: %s", err)
		}
	}

	return nil
}

func (f *Frame) getFormat(data io.ByteReader) error {
	b, err := data.ReadByte()
	if err != nil {
		return err
	}
	f.formatType = b >> 4
	f.length = int(b&0x7) << 8
	if f.formatType != 0xa {
		return fmt.Errorf("Unexpected frame format type: 0x%x", f.formatType)
	}
	b, err = data.ReadByte()
	if err != nil {
		return err
	}
	f.length += int(b)
	return nil
}

func getAddress(data io.ByteReader) (addr []byte, err error) {
	for i := 0; i < 4; i++ {
		var b byte
		b, err = data.ReadByte()
		if err != nil {
			return nil, err
		}
		addr = append(addr, b)
		if b&0x01 == 0x01 {
			return
		}
	}
	return nil, fmt.Errorf("Address end could not be found")
}

func verifyChecksum(data io.ByteReader, expected uint16) error {
	var checksum uint16
	for i := 0; i < 2; i++ {
		b, err := data.ReadByte()
		if err != nil {
			return fmt.Errorf("Failed to read checksum byte #%d: %s", i+1, err)
		}
		checksum += uint16(b) << uint(8*i)
	}
	if checksum != expected {
		return fmt.Errorf("Checksums don't match: expected=0x%04x calculated=0x%04x", expected, checksum)
	}
	return nil
}

func (f *Frame) String() string {
	return fmt.Sprintf("{dest=%v src=%v ctrl=0x%02x info=%s}",
		f.dest, f.src, f.ctrl, hex.EncodeToString(f.info))
}
