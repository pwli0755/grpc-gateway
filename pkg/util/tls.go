package util

import (
	"crypto/tls"
	"golang.org/x/net/http2"
	"io/ioutil"
	"log"
)

func GetTLSConfig(certPemPath, certKeyPath string) (*tls.Config, error) {
	var certKeyPair *tls.Certificate
	cert, err := ioutil.ReadFile(certPemPath)
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(certKeyPath)
	if err != nil {
		return nil, err
	}
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		log.Printf("TLS KeyPair err: %v\n", err)
		return nil, err

	}
	certKeyPair = &pair
	return &tls.Config{
		Certificates: []tls.Certificate{*certKeyPair},
		NextProtos:   []string{http2.NextProtoTLS},
	}, nil
}
