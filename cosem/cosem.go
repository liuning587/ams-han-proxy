package cosem

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	choiceNullData           byte = 0 // nil
	choiceArray                   = 1
	choiceStructure               = 2
	choiceBool                    = 3  // bool
	choiceDoubleLong              = 5  // int32
	choiceDoubleLongUnsigned      = 6  // uint32
	choiceOctetString             = 9  // []byte
	choiceVisibleString           = 10 // string
	choiceUTF8String              = 12 // string
	choiceInteger                 = 15 // int8
	choiceLongInteger             = 16 // int16
	choiceUnsigned                = 17 // uint8
	choiceLongUnsigned            = 18 // uint16
	choiceLong64                  = 20 // int64
	choiceLong64Unsigned          = 21 // uint64
	choiceFloat32                 = 23 // float32
	choiceFloat64                 = 24 // float64
	choiceDateTime                = 25 // time.Time
)

type decodeFunc func(*bytes.Buffer) (Data, error)

var decodeMapper = map[byte]decodeFunc{
	choiceNullData:           decodeNullData,
	choiceDoubleLong:         func(buf *bytes.Buffer) (Data, error) { return decodeInteger(buf, 4, true) },
	choiceDoubleLongUnsigned: func(buf *bytes.Buffer) (Data, error) { return decodeInteger(buf, 4, false) },
	choiceOctetString:        decodeOctetString,
	choiceInteger:            func(buf *bytes.Buffer) (Data, error) { return decodeInteger(buf, 1, true) },
	choiceLongInteger:        func(buf *bytes.Buffer) (Data, error) { return decodeInteger(buf, 2, true) },
	choiceUnsigned:           func(buf *bytes.Buffer) (Data, error) { return decodeInteger(buf, 1, false) },
	choiceLongUnsigned:       func(buf *bytes.Buffer) (Data, error) { return decodeInteger(buf, 2, false) },
	choiceDateTime:           decodeDateTime,
}

func init() {
	decodeMapper[choiceStructure] = decodeStructure
	decodeMapper[choiceArray] = decodeArray
}

type Data interface {
	fmt.Stringer
}

type Array struct{ *Telegram }

type Structure struct{ *Telegram }

type Telegram struct {
	items []Data
}

func DecodeTelegram(data []byte) (*Telegram, error) {
	buf := bytes.NewBuffer(data)
	return decodeTelegram(buf)
}

func (t *Telegram) String() string {
	result := "{"
	for _, i := range t.items {
		result += i.String() + " "
	}
	return strings.TrimRight(result, " ") + "}"
}

func (t *Telegram) NumItems() int {
	return len(t.items)
}

func (t *Telegram) Item(i int) Data {
	return t.items[i]
}

func decodeTelegram(buf *bytes.Buffer) (*Telegram, error) {
	t := &Telegram{}
	for buf.Len() > 0 {
		b, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("Failed to read choice byte: %s", err)
		}
		decoder, ok := decodeMapper[b]
		if !ok {
			return nil, fmt.Errorf("Unexpected choice byte (%d)", b)
		}
		item, err := decoder(buf)
		if err != nil {
			return nil, fmt.Errorf("Decoder failed (choice byte = %d): %s", b, err)
		}
		t.items = append(t.items, item)
	}

	return t, nil
}

func decodeStructure(buf *bytes.Buffer) (Data, error) {
	if _, err := getLength(buf); err != nil {
		return nil, err
	}
	t, err := decodeTelegram(buf)
	return Structure{t}, err
}

func decodeArray(buf *bytes.Buffer) (Data, error) {
	if _, err := getLength(buf); err != nil {
		return nil, err
	}
	t, err := decodeTelegram(buf)
	return Array{t}, err
}

func getLength(buf *bytes.Buffer) (int, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("Failed to read length byte: %s", err)
	}
	return int(b), nil
}
