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

	certFile := fmt.Sprintf("../certs/certs/new-device-full-chain.cert.pem")
	keyFile := fmt.Sprintf("../certs/private/new-device.key.pem")

	clientKeyPair, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		log.Fatalln(err.Error())
	}

	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientKeyPair},
	}
}
