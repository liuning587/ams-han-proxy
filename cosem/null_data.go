package cosem

import "bytes"

type NullData struct{}

func decodeNullData(*bytes.Buffer) (Data, error) {
	return NullData{}, nil
}

func (NullData) String() string {
	return "<nil>"
}
