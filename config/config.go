package config

import (
	"fmt"
	"strings"

	"github.com/jacobsa/go-serial/serial"
)

type Config struct {
	Serial Serial
}

type Serial struct {
	Device   string `required:"true"`
	BaudRate uint   `split_words:"true" default:"2400"`
	DataBits uint   `split_words:"true" default:"8"`
	Parity   parity `default:"even" desc:"even,odd,none"`
	StopBits uint   `split_words:"true" default:"1"`
}

func (s *Serial) GetOpenOptions() serial.OpenOptions {
	return serial.OpenOptions{
		PortName:   s.Device,
		BaudRate:   s.BaudRate,
		DataBits:   s.DataBits,
		ParityMode: s.Parity.ParityMode,
		StopBits:   s.StopBits,
	}
}

type parity struct{ serial.ParityMode }

func (p *parity) Decode(str string) error {
	switch strings.ToLower(str) {
	case "even":
		fallthrough
	case "e":
		p.ParityMode = serial.PARITY_EVEN
	case "odd":
		fallthrough
	case "o":
		p.ParityMode = serial.PARITY_ODD
	case "none":
		fallthrough
	case "n":
		p.ParityMode = serial.PARITY_NONE
	default:
		return fmt.Errorf("Unknown parity: %s", str)
	}
	return nil
}
