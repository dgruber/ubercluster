/*
   Copyright 2017 Daniel Gruber

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"fmt"
	"github.com/dgruber/ubercluster/pkg/persistency"
	"github.com/dgruber/ubercluster/pkg/proxy"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"log"
	"os"
)

var verbose = false

func init() {
	if verbose == false {
		log.SetOutput(ioutil.Discard)
	}
}

// Standard set of CLI parameters.
var (
	app            = kingpin.New("dockerproxy", "A proxy server for Docker")
	cliVerbose     = app.Flag("verbose", "Enables enhanced logging for debugging.").Bool()
	cliPort        = app.Flag("port", "Sets address and port on which proxy is listening.").Default(":8080").String()
	certFile       = app.Flag("certFile", "Path to certification file for secure connections (TLS).").Default("").String()
	keyFile        = app.Flag("keyFile", "Path to key file for secure connections (TLS).").Default("").String()
	otp            = app.Flag("otp", "One time password settings (\"yubikey\") or a fixed shared secret.").Default("").String()
	yubiID         = app.Flag("yubiID", "Yubi client ID if otp is set to yubikey.").Default("").String()
	yubiSecret     = app.Flag("yubiSecret", "Yubi secret key if otp is set to yubikey").Default("").String()
	yubiAllowedIds = app.Flag("yubiAllowedIds", "A list of IDs of yubikeys which are accepted as source for OTPs.").Default("").Strings()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *cliVerbose {
		log.SetOutput(os.Stdout)
	}

	cf, err := New()
	if err != nil {
		fmt.Printf("Error during initialization: %s\n", err)
		os.Exit(1)
	}

	var sc proxy.SecConfig
	sc.OTP = *otp
	sc.YubiID = *yubiID
	sc.YubiSecret = *yubiSecret
	sc.YubiAllowedIDs = *yubiAllowedIds

	var ps persistency.DummyPersistency

	proxy.ProxyListenAndServe(*cliPort, *certFile, *keyFile, sc, &ps, cf)
}
