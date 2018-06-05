package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"os"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/jacobsa/go-serial/serial"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"github.com/wercker/journalhook"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"svenschwermer.de/ams-han-proxy/client/config"
	"svenschwermer.de/ams-han-proxy/han"
	"svenschwermer.de/ams-han-proxy/hdlc"
	api "svenschwermer.de/ams-han-proxy/proto/electricity"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
}

func main() {
	journalhook.Enable()
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
	cc, err := grpc.Dial(cfg.GRPC.Address, dialOption,
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithMax(5),
			grpc_retry.WithPerRetryTimeout(3*time.Second),
			grpc_retry.WithCodes(append(grpc_retry.DefaultRetriableCodes, codes.Internal)...),
		)))
	if err != nil {
		log.Fatal(err)
	}
	sink := api.NewMeterSinkClient(cc)

	publishJobs := make(chan *api.MeterData)
	defer close(publishJobs)
	for i := 0; i < 20; i++ {
		go func() {
			for md := range publishJobs {
				log.Debugf("Publishing %+v", *md)
				_, err = sink.Publish(context.Background(), md)
				if err != nil {
					log.Errorf("Failed to publish meter data (%+v): %s", *md, err)
				}
			}
		}()
	}

	scanner := bufio.NewScanner(port)
	scanner.Split(hdlc.SplitFrames)
	for scanner.Scan() {
		f := &hdlc.Frame{}
		if err := f.UnmarshalBinary(scanner.Bytes()); err != nil {
			log.Errorf("Failed to decode HDLC frame (%s): %s",
				hex.EncodeToString(scanner.Bytes()), err)
		} else {
			md, err := han.DecodeLLCPayload(f.LogicalLinkLayerPayload())
			if err != nil {
				log.Errorf("Failed to decode LLC payload (%s): %s",
					hex.EncodeToString(f.LogicalLinkLayerPayload()), err)
			} else {
				publishJobs <- md
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
