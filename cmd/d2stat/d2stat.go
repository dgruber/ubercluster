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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dgruber/drmaa2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// configuration for proxies of compute clusters which can be queried
var config Config

type ClusterConfig struct {
	// name to reference the cluster in this tool ("default" is 
	// the address used when no cluster is explicitly referenced
	Name string
	// like: http://myhost:8888
	Address string
}

// Complete configuration
type Config struct {
	// Multiple endpoints of proxies can be defined
	Cluster []ClusterConfig
}

// saveDummyConfig creates a file dummyconfig.json in order
// to help the user to create a configuration file (config.json)
// in the right JSON format. "default" and "cluster1" are 
// user defined names for the cluster, while the address is 
// the endpoint of the proxy.
func saveDummyConfig() {
	if file, err := os.Create("dummyconfig.json"); err == nil {
		encoder := json.NewEncoder(file)
		var config Config
		config.Cluster = make([]ClusterConfig, 0)
		var def, cluster ClusterConfig
		def.Name = "default"
		def.Address = "http://localhost:8888/"
		cluster.Name = "cluster1"
		cluster.Address = "http://localhost:8282/"
		config.Cluster = append(config.Cluster, def)
		config.Cluster = append(config.Cluster, cluster)
		encoder.Encode(config)
		file.Close()
	}
}

// Disable logging by default
func init() {
	log.SetOutput(ioutil.Discard)
}

func makeDate(date time.Time) string {
	if date.Unix() == drmaa2.UnsetTime {
		return "-"
	}
	if date.Unix() == drmaa2.InfiniteTime {
		return "inf"
	}
	if date.Unix() == drmaa2.ZeroTime {
		return "0"
	}
	return date.String()
}

// emulateQstat prints DRMAA2 JobInfo information on
// stdout in a similar way than qstat -j (same keyes)
func emulateQstat(ji drmaa2.JobInfo) {
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
func emulateQhost(m drmaa2.Machine) {
	fmt.Fprintf(os.Stdout, "%s %s %d %d %d %f %d %d\n", m.Name, m.Architecture.String(), m.Sockets,
		m.Sockets*m.CoresPerSocket, m.Sockets*m.CoresPerSocket*m.ThreadsPerCore, m.Load,
		m.PhysicalMemory, m.VirtualMemory)
}

func showJobDetails(clustername, jobid string) {
	request := fmt.Sprintf("%s%s%s", clustername, "/monitoring?jobid=", jobid)
	log.Println("Requesting:" + request)
	resp, err := http.Get(request)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var jobinfo drmaa2.JobInfo
	if err := decoder.Decode(&jobinfo); err == nil {
		// here formating rules
		emulateQstat(jobinfo)
	}
}

func showJobsInState(clustername, state string) {
	request := fmt.Sprintf("%s%s%s", clustername, "/monitoring?state=", state)
	log.Println("Requesting:" + request)
	resp, err := http.Get(request)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var joblist []drmaa2.JobInfo
	decoder.Decode(&joblist)
	// here formating rules
	for index, _ := range joblist {
		emulateQstat(joblist[index])
		fmt.Println()
	}
	if len(joblist) == 0 {
		fmt.Printf("No job in state %s found.\n", state)
	}
}

// submitJob creates a new job in the given cluster
func submitJob(clustername, jobname, cmd, arg, queue string) {
	var jt drmaa2.JobTemplate
	// fill a DRMAA2 job template and send it over to the proxy
	jt.RemoteCommand = cmd
	jt.JobName = jobname
	if arg != "" {
		jt.Args = []string{arg}
	}
	jt.QueueName = queue
	jtb, _ := json.Marshal(jt)

	// create URL of cluster to send the job to
	url := fmt.Sprintf("%s%s", clustername, "/session")

	log.Println("Submit template: ", string(jtb))
	if resp, err := http.Post(url, "application/json", bytes.NewBuffer(jtb)); err != nil {
		fmt.Println("Error during post: ", err)
	} else {
		log.Println("Status of request:", resp.Status)
	}
}

func showQueues(clustername, queue string) {
	showMachinesQueues(clustername, "queues", queue)
}

func showMachines(clustername, machine string) {
	showMachinesQueues(clustername, "machines", machine)
}

func showMachinesQueues(clustername, req, filter string) {
	request := fmt.Sprintf("%s/monitoring?%s=%s", clustername, req, filter)
	log.Println("Requesting:" + request)
	resp, err := http.Get(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	if req == "machines" {
		var machinelist []drmaa2.Machine
		if err := decoder.Decode(&machinelist); err != nil {
			fmt.Println("Error during decoding: ", err)
			os.Exit(1)
		}
		for index, _ := range machinelist {
			emulateQhost(machinelist[index])
		}
	} else if req == "queues" {
		var queuelist []drmaa2.Queue
		if err := decoder.Decode(&queuelist); err != nil {
			fmt.Println("Error during decoding: ", err)
			os.Exit(1)
		}
		for index, _ := range queuelist {
			fmt.Println(queuelist[index].Name)
		}
	}
}

func main() {
	f := flag.NewFlagSet("d2stat", flag.ExitOnError)
	verbose := f.String("v", "", "Turns on logging for debugging (v).")
	jobid := f.String("j", "0", "Displays information about a particular job or \"all\"")
	state := f.String("s", "", "Show all jobs with a certain state (r/f/p..)")
	machine := f.String("m", "", "Shows details of the compute machine of the cluster")
	queue := f.String("q", "", "Lists all available queues from the cluster scheduler")
	cluster := f.String("c", "default", "Defines the cluster name on which the operation is performed on.")
	submit := f.String("submit", "", "Submits that job into the given cluster.")
	arg := f.String("arg", "", "Argument of the command which is submitted.")
	name := f.String("name", "", "Name of the job which is submitted.")
	q := f.String("queue", "", "Queue name when submitting job.")
	// ... JSDL?
	saveDummyConfig()

	if len(os.Args) <= 1 {
		fmt.Println("Unknown arguments. Try -help ...")
		os.Exit(2)
		return
	}

	if *verbose != "" {
		log.SetOutput(os.Stdout)
	}

	// read in configuration
	if file, err := os.Open("config.json"); err != nil {
		fmt.Println("Can't read configuration (config.json) file.")
		os.Exit(1)
	} else {
		decoder := json.NewDecoder(file)
		decoder.Decode(&config)
		log.Println(config)
	}

	// parse command line
	if err := f.Parse(os.Args[1:]); err != nil {
		fmt.Println("Error during parsing: ", err)
		os.Exit(2)
	}

	// select cluster
	var clusteraddress string
	for i, _ := range config.Cluster {
		if *cluster == config.Cluster[i].Name {
			clusteraddress = config.Cluster[i].Address
		}
	}
	if clusteraddress == "" {
		fmt.Println("Cluster name %s not found in configuration.", *cluster)
		os.Exit(1)
	}

	if *jobid != "0" {
		showJobDetails(clusteraddress, *jobid)
	} else if *state != "" {
		showJobsInState(clusteraddress, *state)
	} else if *queue != "" {
		showQueues(clusteraddress, *queue)
	} else if *submit != "" {
		submitJob(clusteraddress, *name, *submit, *arg, *q)
	} else if *machine != "" {
		fmt.Fprintf(os.Stdout, "HOSTNAME ARCH NSOC NCOR NTHR LOAD MEMTOT SWAPTO\n")
		showMachines(clusteraddress, *machine)
	} else {
		f.Usage()
	}
}
