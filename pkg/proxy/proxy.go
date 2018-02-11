package proxy

import (
	"crypto/tls"
	"fmt"
	"github.com/dgruber/ubercluster/pkg/persistency"
	"net/http"
	"os"
)

// ProxyListenAndServe starts an http proxy for a cluster which is accessed by functions
// specified in the ProxyImplementer interface. If a certification and key file is given
// as parameter then it starts an TLS secured http proxy. The port is specified by addr
// in the form which is used by http.ListenAndServe.
func ProxyListenAndServe(addr, certFile, keyFile string, sc SecConfig, pi persistency.PersistencyImplementer, impl ProxyImplementer) {
	if certFile != "" && keyFile != "" {

		clientCertPool, _ := ReadTrustedClientCertPool(sc.TrustedClientCertDir)

		servTLSCert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			fmt.Printf("invalid key pair: %v\n", err)
			os.Exit(1)
		}

		tlsConfig := &tls.Config{
			// Reject any TLS certificate that cannot be validated
			ClientAuth: tls.RequireAndVerifyClientCert,
			// Ensure that we only use our "CA" to validate certificates
			ClientCAs:    clientCertPool,
			Certificates: []tls.Certificate{servTLSCert},
			//CipherSuites:             []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},
			//PreferServerCipherSuites: true,
			//MinVersion: tls.VersionTLS12,
			//InsecureSkipVerify: true,
		}

		tlsConfig.BuildNameToCertificate()

		httpServer := &http.Server{
			Addr:      addr,
			TLSConfig: tlsConfig,
			Handler:   NewProxyRouter(impl, sc, pi),
		}
		if err := httpServer.ListenAndServeTLS(certFile, keyFile); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		fmt.Println("starting plain http server")
		if err := http.ListenAndServe(addr, NewProxyRouter(impl, sc, pi)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
