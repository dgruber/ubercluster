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
	"encoding/json"
	"fmt"
	"github.com/dgruber/ubercluster/pkg/types"
	"io"
	"log"
)

// JSONFormat defines how information is published.
type JSONFormat struct {
	output io.Writer // defines where to print
}

func (jf *JSONFormat) marshalJSON(data interface{}) {
	if out, err := json.Marshal(data); err != nil {
		log.Panic(err)
	} else {
		fmt.Fprintf(jf.output, "%s", string(out))
	}
}

// PrintFiles writes information about each file in one
// line in the configured output stream
func (jf *JSONFormat) PrintFiles(fs []types.FileInfo) {
	jf.marshalJSON(fs)
}

func (jf *JSONFormat) PrintJobDetails(ji types.JobInfo) {
	jf.marshalJSON(ji)
}

func (jf *JSONFormat) PrintMachine(m types.Machine) {
	jf.marshalJSON(m)
}
