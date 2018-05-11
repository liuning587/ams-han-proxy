package han

import (
	"encoding/hex"
	"log"

	"svenschwermer.de/ams-han-proxy/cosem"
)

func DecodeKFM001(data []byte) {
	t, err := cosem.DecodeTelegram(data)
	if err != nil {
		log.Printf("Failed to decode telegram (%s): %s", hex.EncodeToString(data), err)
	} else {
		if t.NumItems() != 6 {
			log.Printf("Expected 6 items, got %d: %v", t.NumItems(), t)
		} else {
			d := t.Item(5)
			switch s := d.(type) {
			case cosem.Structure:
				switch s.NumItems() {
				case 1:
					HandleKFM001List1(s)
				case 9:
					HandleKFM001List2(s)
				case 14:
					HandleKFM001List3(s)
				default:
					log.Printf("Unexpected structure in telegram (%v): %v", t, s)
				}
			default:
				log.Printf("Expected structure, got %T (Telegram: %v)", d, t)
			}
		}
	}
}

func HandleKFM001List1(s cosem.Structure) error {
	log.Printf("Active power+: %d W", s.Item(0).(cosem.Integer))
	return nil
}

func HandleKFM001List2(s cosem.Structure) error {
	log.Printf("OBIS List version identifier: %s", string(s.Item(0).(cosem.OctetString)))
	log.Printf("Meter ID: %s", string(s.Item(1).(cosem.OctetString)))
	log.Printf("Meter type: %s", string(s.Item(2).(cosem.OctetString)))
	log.Printf("Active power+: %d W", s.Item(3).(cosem.Integer))
	log.Printf("Active power-: %d W", s.Item(4).(cosem.Integer))
	log.Printf("Reactive power+: %d VAr", s.Item(5).(cosem.Integer))
	log.Printf("Reactive power-: %d VAr", s.Item(6).(cosem.Integer))
	log.Printf("Phase current (?): %.3f A", float64(s.Item(7).(cosem.Integer))/1000)
	log.Printf("Phase voltage (?): %.1f V", float64(s.Item(8).(cosem.Integer))/10)
	return nil
}

func HandleKFM001List3(s cosem.Structure) error {
	if err := HandleKFM001List2(s); err != nil {
		return err
	}
	t, err := s.Item(9).(cosem.OctetString).AsDateTime()
	if err != nil {
		return err
	}
	log.Printf("Meter time: %v", t)
	// 9: Timestamp = 07e2050a0417000aff800000
	// 10: Cumulative hourly active import energy (A+) (Q1+Q4) = 1436570
	// 11: Cumulative hourly active export energy (A-) (Q2+Q3) = 0
	// 12: Cumulative hourly reactive import energy (R+) (Q1+Q2) = 40936
	// 13: Cumulative hourly reactive export energy (R-) (Q3+Q4) = 53402
	return nil
}
