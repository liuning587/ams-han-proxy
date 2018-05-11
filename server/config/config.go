package config

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	influx "github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	LogLevel logLevel `split_words:"true" default:"INFO"`
	GRPC     GRPC
	Influx   Influx
}

type logLevel struct{ log.Level }

func (l *logLevel) Decode(str string) (err error) {
	l.Level, err = log.ParseLevel(str)
	return
}

type GRPC struct {
	ListenAddress string `split_words:"true" required:"true"`
	Certificate   string
	PrivateKey    string   `split_words:"true"`
	ClientCA      []string `split_words:"true"`
}

func (g *GRPC) GetCredentials() (credentials.TransportCredentials, error) {
	if g.Certificate == "" && g.PrivateKey == "" {
		log.Warn("Listening for an insecure connection")
		return nil, nil
	}
	cert, err := tls.LoadX509KeyPair(g.Certificate, g.PrivateKey)
	if err != nil {
		return nil, err // TODO: give additional info
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    x509.NewCertPool(),
	}
	for _, c := range g.ClientCA {
		bytes, err := ioutil.ReadFile(c)
		if err != nil {
			return nil, err // TODO: give additional info
		}
		tlsConfig.ClientCAs.AppendCertsFromPEM(bytes)
	}
	return credentials.NewTLS(tlsConfig), nil
}

type Influx struct {
	Address  string `required:"true"`
	Database string `required:"true"`
}

func (i *Influx) GetHTTPConfig() influx.HTTPConfig {
	return influx.HTTPConfig{Addr: i.Address}
}
