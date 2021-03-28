package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
)

var brokerCACert = "../certs/IoTHubRootCA_Baltimore.pem"

func newTlsConfig() *tls.Config {

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(brokerCACert)
	if err != nil {
		log.Fatalln(err.Error())
	}

	certPool.AppendCertsFromPEM(ca)

	certFile := fmt.Sprintf("../certs/%s-public.pem", ctx.ClientID)
	keyFile := fmt.Sprintf("../certs/%s-private.pem", ctx.ClientID)

	clientKeyPair, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		log.Fatalln(err.Error())
	}

	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientKeyPair},
	}
}
