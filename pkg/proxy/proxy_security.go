package proxy

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
)

// SecConfig stores security related configuration settings for the ubercluster Proxy
type SecConfig struct {
	OTP                  string   // secret key or "yubikey"
	YubiID               string   // ID of yubiservice in case of yubikey https://upgrade.yubico.com/getapikey/
	YubiSecret           string   // Secret of yubiservice in case of yubikey https://upgrade.yubico.com/getapikey/
	YubiAllowedIDs       []string // IDs of yubkeys which are allowed
	TrustedClientCertDir string   // Directory which contains trusted certs for mutual TLS
}

func ReadTrustedClientCertPool(directory string) (*x509.CertPool, error) {
	fileinfos, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	clientCertPool := x509.NewCertPool()

	for i := range fileinfos {
		file := directory + string(os.PathSeparator) + fileinfos[i].Name()

		// added trusted client certs
		certBytes, err := ioutil.ReadFile(file)
		if err != nil {
			return clientCertPool, err
		}

		if ok := clientCertPool.AppendCertsFromPEM(certBytes); !ok {
			return clientCertPool, fmt.Errorf("unable to add certificate %s to certificate pool", file)
		}
	}

	return clientCertPool, nil
}
