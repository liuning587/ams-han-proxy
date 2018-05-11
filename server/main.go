package main

import (
	"net"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	api "svenschwermer.de/ams-han-proxy/proto/electricity"
	"svenschwermer.de/ams-han-proxy/server/config"
	"svenschwermer.de/ams-han-proxy/server/electricity"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
}

func main() {
	cfg := &config.Config{}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatal(err)
	}

	log.SetLevel(cfg.LogLevel.Level)

	creds, err := cfg.GRPC.GetCredentials()
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer(grpc.Creds(creds))

	handler, err := electricity.NewHandler(cfg.Influx.GetHTTPConfig(), cfg.Influx.Database)
	if err != nil {
		log.Fatal(err)
	}
	api.RegisterMeterSinkServer(server, handler)

	listener, err := net.Listen("tcp", cfg.GRPC.ListenAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Listening on %v", listener.Addr())
	log.Fatal(server.Serve(listener))
}
