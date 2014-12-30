/*
   Copyright 2014 Daniel Gruber, Univa

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
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var verbose bool = false
var JobSessionName = "ubercluster"

func init() {
	if verbose == false {
		log.SetOutput(ioutil.Discard)
	}
}

var (
	app        = kingpin.New("d2proxy", "A proxy server for DRMAA2 compatible cluster schedulers (like Univa Grid Engine).")
	cliVerbose = app.Flag("verbose", "Enables enhanced logging for debugging.").Bool()
	cliPort    = app.Flag("port", "Sets address and port on which proxy is listening.").Default(":8888").String()
	certFile   = app.Flag("certFile", "Path to certification file for secure connections (TLS).").Default("").String()
	keyFile    = app.Flag("keyFile", "Path to key file for secure connections (TLS).").Default("").String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *cliVerbose {
		log.SetOutput(os.Stdout)
	}

	// read-in config
	initializeD2Proxy()

	// Open MonitoringSession and create a JobSession with the given name
	initializeDRMAA2(JobSessionName)

	if *certFile != "" && *keyFile != "" {
		if err := http.ListenAndServeTLS(*cliPort, *certFile, *keyFile, NewProxyRouter()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServe(*cliPort, NewProxyRouter()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
