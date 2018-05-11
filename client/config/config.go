package config

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/credentials"

	"github.com/jacobsa/go-serial/serial"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Config struct {
	LogLevel logLevel `split_words:"true" default:"INFO"`
	Serial   Serial
	GRPC     GRPC
}

type logLevel struct{ log.Level }

func (l *logLevel) Decode(str string) (err error) {
	l.Level, err = log.ParseLevel(str)
	return
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

type GRPC struct {
	Address     string `required:"true"`
	Certificate string
	PrivateKey  string `split_words:"true"`
}

func (g *GRPC) GetDialOption() (grpc.DialOption, error) {
	if g.Certificate == "" && g.PrivateKey == "" {
		log.Warn("Establishing insecure connection")
		return grpc.WithInsecure(), nil
	}
	creds, err := credentials.NewServerTLSFromFile(g.Certificate, g.PrivateKey)
	if err != nil {
		return nil, err
	}
	return grpc.WithTransportCredentials(creds), nil
}
