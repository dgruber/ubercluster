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
	"io"
	"os"
	"time"
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

func makeDate(date time.Time) string {
	if date.Unix() == types.UnsetTime {
		return "-"
	}
	if date.Unix() == types.InfiniteTime {
		return "inf"
	}
	if date.Unix() == types.ZeroTime {
		return "0"
	}
	return date.String()
}

// emulateQstat prints DRMAA2 JobInfo information on
// stdout in a similar way than qstat -j (same keyes)
func emulateQstat(ji types.JobInfo) {
	fmt.Fprintf(os.Stdout, "job_number:\t\t%s\n", ji.Id)
	fmt.Fprintf(os.Stdout, "state:\t\t\t%s\n", ji.State)
	fmt.Fprintf(os.Stdout, "submission_time:\t%s\n", makeDate(ji.SubmissionTime))
	fmt.Fprintf(os.Stdout, "dispatch_time:\t\t%s\n", makeDate(ji.DispatchTime))
	fmt.Fprintf(os.Stdout, "finish_time:\t\t%s\n", makeDate(ji.FinishTime))
	fmt.Fprintf(os.Stdout, "owner:\t\t\t%s\n", ji.JobOwner)
	fmt.Fprintf(os.Stdout, "slots:\t\t\t%d\n", ji.Slots)
	fmt.Fprintf(os.Stdout, "allocated_machines:\t")
	if ji.AllocatedMachines != nil {
		first := true
		for _, machine := range ji.AllocatedMachines {
			if machine != "" {
				if first {
					first = false
					fmt.Fprintf(os.Stdout, "%s", machine)
				} else {
					fmt.Fprintf(os.Stdout, ",%s", machine)
				}
			}
		}
		fmt.Fprintf(os.Stdout, "\n")
	} else {
		fmt.Fprintf(os.Stdout, "NONE\n")
	}
	fmt.Fprintf(os.Stdout, "exit_status:\t\t%d\n", ji.ExitStatus)
}

// emulateQhost prints machine information in SGE style out
func emulateQhost(m types.Machine) {
	fmt.Fprintf(os.Stdout, "%s %s %d %d %d %f %d %d\n", m.Name, m.Architecture.String(), m.Sockets,
		m.Sockets*m.CoresPerSocket, m.Sockets*m.CoresPerSocket*m.ThreadsPerCore, m.Load,
		m.PhysicalMemory, m.VirtualMemory)
}

func (sf *StandardFormat) PrintJobDetails(ji types.JobInfo) {
	emulateQstat(ji)
}

func (sf *StandardFormat) PrintMachine(m types.Machine) {
	emulateQhost(m)
}
