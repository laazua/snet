package snet

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
)

var (
	serverAuthConfig *tls.Config
	clientAuthConfig *tls.Config
)

func load(cafile, certfile, keyfile string) (tls.Certificate, *x509.CertPool, error) {
	cert, err := tls.LoadX509KeyPair(certfile, keyfile)
	if err != nil {
		log.Fatal(err)
		return tls.Certificate{}, nil, err
	}

	caCert, err := os.ReadFile(cafile)
	if err != nil {
		log.Fatal(err)
		return tls.Certificate{}, nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	return cert, certPool, nil
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		// 其他错误（如权限问题）
		return false
	}
}

// 设置服务器端认证配置
func SetServerAuth(caFile, crtFile, keyFile string) error {
	if !fileExists(crtFile) || !fileExists(keyFile) || !fileExists(caFile) {
		return ErrCertFileNotFound
	}
	cert, certPool, err := load(caFile, crtFile, keyFile)
	if err != nil {
		return err
	}
	serverAuthConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}
	return nil
}

// 设置客户端认证配置
func SetClientAuth(caFile, crtFile, keyFile string) error {
	if !fileExists(crtFile) || !fileExists(keyFile) || !fileExists(caFile) {
		return ErrCertFileNotFound
	}
	cert, certPool, err := load(caFile, crtFile, keyFile)
	if err != nil {
		return err
	}
	clientAuthConfig = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            certPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}
	return nil
}
