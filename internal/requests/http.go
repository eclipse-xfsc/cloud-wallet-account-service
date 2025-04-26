package requests

import (
	"crypto/tls"
	"crypto/x509"
	cmn "github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"net/http"
	"os"
)

func httpTLSClient(config http.Client) (*http.Client, error) {
	certFile := cmn.GetEnv("TLS_CERT_FILE", "cert.pem")
	keyFile := cmn.GetEnv("TLS_KEY_FILE", "key.pem")

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Create a CA certificate pool and add cert.pem to it
	caCert, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create a HTTPS client and supply the created CA pool and certificate
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
		CheckRedirect: config.CheckRedirect,
		Jar:           config.Jar,
		Timeout:       config.Timeout,
	}

	return client, nil
}

func HttpClient(withTLS bool, config http.Client) (*http.Client, error) {
	if withTLS {
		return httpTLSClient(config)
	} else {
		return &http.Client{
			Transport:     config.Transport,
			CheckRedirect: config.CheckRedirect,
			Jar:           config.Jar,
			Timeout:       config.Timeout,
		}, nil
	}
}
