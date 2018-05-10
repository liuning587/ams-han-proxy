package cosem

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type OctetString []byte

func decodeOctetString(buf *bytes.Buffer) (Data, error) {
	l, err := getLength(buf)
	if err != nil {
		return nil, err
	}
	bytes := make(OctetString, l)
	n, err := buf.Read(bytes)
	if n != l || err != nil {
		return nil, fmt.Errorf("Failed to read bytes from buffer")
	}
	return bytes, nil
}

func (os OctetString) String() string {
	return hex.EncodeToString(os)
}
