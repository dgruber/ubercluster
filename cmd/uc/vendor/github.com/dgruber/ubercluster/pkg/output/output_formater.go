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

package output

import (
	"fmt"
	"github.com/dgruber/ubercluster/pkg/types"
	"log"
	"os"
)

// OutputFormater is an interface which defines
// all required functions needed for the uc client
// to print out the results.
type OutputFormater interface {
	PrintFiles(fs []types.FileInfo) // output format of "uc ls"
	PrintJobDetails(ji types.JobInfo)
	PrintMachine(m types.Machine)
}

// MakeOutputFormater creates an output formater depending
// on the chosen output format.
func MakeOutputFormater(format string) OutputFormater {
	switch format {
	case "default":
		log.Println("Standard output format selected.")
		var sf StandardFormat
		sf.output = os.Stdout
		return &sf
	case "JSON", "json":
		log.Println("JSON output format selected.")
		var jf JSONFormat
		jf.output = os.Stdout
		return &jf
	case "XML", "xml":
		log.Println("XML output format selected.")
		var jf XMLFormat
		jf.output = os.Stdout
		return &jf
	}
	fmt.Println("Error selecting output format module.")
	os.Exit(1)
	return nil
}
