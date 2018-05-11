package main

import (
	"bufio"

	"github.com/jacobsa/go-serial/serial"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"

	"svenschwermer.de/ams-han-proxy/config"
	"svenschwermer.de/ams-han-proxy/han"
	"svenschwermer.de/ams-han-proxy/hdlc"
)

func main() {
	cfg := &config.Config{}
	if err := envconfig.Process("", cfg); err != nil {
		envconfig.Usage("", cfg)
		log.Fatal(err)
	}

	serialOptions := cfg.Serial.GetOpenOptions()
	serialOptions.MinimumReadSize = 1
	port, err := serial.Open(serialOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	scanner := bufio.NewScanner(port)
	scanner.Split(hdlc.SplitFrames)
	for scanner.Scan() {
		f := &hdlc.Frame{}
		if err := f.UnmarshalBinary(scanner.Bytes()); err != nil {
			log.Errorf("Failed to decode frame: %s", err)
		} else {
			han.DecodeKFM001(f.LogicalLinkLayerPayload())
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
