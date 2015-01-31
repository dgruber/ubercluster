/*
   Copyright 2015 Daniel Gruber, Univa, My blog: www.gridengine.eu

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

package ubercluster

import (
	"fmt"
	"io"
	"log"
	"os"
)

// OutputFormater is an interface which defines
// all rquired functions needed for the uc client
// to print out the results.
type OutputFormater interface {
	PrintFiles(fs []FileInfo)
}

// MakeOutputFormater creates an output formater depending
// on the chosen output format.
func MakeOutputFormater(format string) OutputFormater {
	switch format {
	case "default":
		log.Println("Standard output format selected.")
		var sf standardFormat
		sf.output = os.Stdout
		return &sf
	}
	fmt.Println("Error selecting output format module.")
	os.Exit(1)
	return nil
}

// Standard implementation
type standardFormat struct {
	output io.Writer
}

// PrintFiles writes information about each file in one
// line in the configured output stream
func (sf *standardFormat) PrintFiles(fs []FileInfo) {
	for i, f := range fs {
		if i != 0 {
			fmt.Fprintln(sf.output)
		}
		kb := f.bytes
		if kb != 0 {
			kb /= 1024
		}
		var exec string
		if f.executable == true {
			exec = "readable"
		} else {
			exec = "executable"
		}
		fmt.Fprintf(sf.output, "%128s %16dkb %s", f.filename, kb, exec)
	}
}
