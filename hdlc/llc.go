package hdlc

import (
	"encoding/hex"
	"fmt"
)

type LogicalLinkLayerPDU struct {
	destinationLSAP uint8 // remote, DLMS/COSEM: 0xe6
	sourceLSAP      uint8 // local, DLMS/COSEM: 0xe6 or 0xe7
	control         uint8 // DLMS/COSEM: LLC_Quality
	Info            []byte
}

func (pdu *LogicalLinkLayerPDU) UnmarshalBinary(data []byte) error {
	n := len(data)
	if n < 3 {
		return fmt.Errorf("Data too short (%d)", n)
	}
	pdu.destinationLSAP = data[0]
	if pdu.destinationLSAP != 0xe6 {
		return fmt.Errorf("Destination LSAP invalid (0x%02x)", pdu.destinationLSAP)
	}
	pdu.sourceLSAP = data[1]
	if (pdu.sourceLSAP & 0xfe) != 0xe6 {
		return fmt.Errorf("Source LSAP invalid (0x%02x)", pdu.sourceLSAP)
	}
	pdu.control = data[2]
	if pdu.control != 0x00 {
		return fmt.Errorf("Control field (LLC_Quality) invalid (0x%02x)", pdu.control)
	}
	pdu.Info = make([]byte, n-3)
	copy(pdu.Info, data[3:n])
	return nil
}

// IsResponse returns whether the LLC PDU is a response or a command
func (pdu *LogicalLinkLayerPDU) IsResponse() bool {
	return (pdu.sourceLSAP & 0x01) == 0x01
}

func (pdu *LogicalLinkLayerPDU) String() string {
	return fmt.Sprintf("{dest=0x%02x src=0x%02x response=%t info=%s}",
		pdu.destinationLSAP, pdu.sourceLSAP, pdu.IsResponse(), hex.EncodeToString(pdu.Info))
}
