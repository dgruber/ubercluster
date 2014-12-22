/*
   Copyright 2014 Daniel Gruber

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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var verbose bool = true

func init() {
	if verbose == false {
		log.SetOutput(ioutil.Discard)
	}
}

func main() {
	// Port number where proxy listens
	f := flag.NewFlagSet("d2proxy", flag.ExitOnError)
	p := f.String("port", ":8888", "Sets address and port on which proxy is listening. (default :8888)")

	if len(os.Args) > 1 {
		if err := f.Parse(os.Args[1:]); err != nil {
			fmt.Println("Error during parsing: ", err)
			os.Exit(2)
		}
	}

	// Open MonitoringSession and create a JobSession with the given name
	initializeDRMAA2("proxy_jsession")

	// Start Proxy
	if err := http.ListenAndServe(*p, NewProxyRouter()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
