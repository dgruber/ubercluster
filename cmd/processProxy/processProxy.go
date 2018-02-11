package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/dgruber/ubercluster/pkg/persistency"
	"github.com/dgruber/ubercluster/pkg/proxy"
	"gopkg.in/alecthomas/kingpin.v1"
)

const SESSION_NAME = "PROCESS_MANAGER"

var verbose = false

func init() {
	if verbose == false {
		log.SetOutput(ioutil.Discard)
	}
}

// Standard set of CLI parameters.
var (
	app                = kingpin.New("processProxy", "An uber-cluster proxy server for managing processes remotely.")
	cliVerbose         = app.Flag("verbose", "Enables enhanced logging for debugging.").Bool()
	cliPort            = app.Flag("port", "Sets address and port on which proxy is listening.").Default(":8888").String()
	certFile           = app.Flag("cert", "Path to certification file for secure connections (TLS).").Default("").String()
	keyFile            = app.Flag("key", "Path to key file for secure connections (TLS).").Default("").String()
	otp                = app.Flag("otp", "One time password settings (\"yubikey\") or a fixed shared secret.").Default("").String()
	trustedClientCerts = app.Flag("clientCerts", "Path to directory where trusted client certificates are stored.").Default("").String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *cliVerbose {
		log.SetOutput(os.Stdout)
	}

	processProxy := NewProxy()
	sc := proxy.SecConfig{
		OTP:                  *otp,
		TrustedClientCertDir: *trustedClientCerts,
	}
	var ps persistency.DummyPersistency

	proxy.ProxyListenAndServe(*cliPort, *certFile, *keyFile, sc, &ps, &processProxy)
}
