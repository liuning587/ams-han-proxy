package main

import (
	"bufio"
	"os"

	"github.com/jacobsa/go-serial/serial"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"svenschwermer.de/ams-han-proxy/config"
	"svenschwermer.de/ams-han-proxy/han"
	"svenschwermer.de/ams-han-proxy/hdlc"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
}

func main() {
	cfg := &config.Config{}
	for _, a := range os.Args {
		if a == "-h" || a == "--help" {
			envconfig.Usage("", cfg)
			os.Exit(0)
		}
	}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatal(err)
	}

	log.SetLevel(cfg.LogLevel.Level)

	serialOptions := cfg.Serial.GetOpenOptions()
	serialOptions.MinimumReadSize = 1
	port, err := serial.Open(serialOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	dialOption, err := cfg.GRPC.GetDialOption()
	if err != nil {
		log.Fatal(err)
	}
	cc, err := grpc.Dial(cfg.GRPC.Address, dialOption)
	if err != nil {
		log.Fatal(err)
	}
	hanHandler := han.NewHandler(cc)

	scanner := bufio.NewScanner(port)
	scanner.Split(hdlc.SplitFrames)
	for scanner.Scan() {
		f := &hdlc.Frame{}
		if err := f.UnmarshalBinary(scanner.Bytes()); err != nil {
			log.Errorf("Failed to decode HDLC frame: %s", err)
		} else {
			err = hanHandler.DecodeLLCPayload(f.LogicalLinkLayerPayload())
			if err != nil {
				log.Warn(err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
