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
	"io"
   "github.com/dgruber/ubercluster/pkg/types"
)

// StandardFormat defines how information is published.
type StandardFormat struct {
	output io.Writer // defines where to print
}

// PrintFiles writes information about each file in one
// line in the configured output stream
func (sf *StandardFormat) PrintFiles(fs []types.FileInfo) {
	for _, f := range fs {
		kb := f.Bytes
		if kb != 0 {
			kb /= 1024
		}
		var exec string
		if f.Executable == false {
			exec = "readable"
		} else {
			exec = "executable"
		}
		fmt.Fprintf(sf.output, "%-40s %12dkb %s", f.Filename, kb, exec)
		fmt.Fprintln(sf.output)
	}
}


