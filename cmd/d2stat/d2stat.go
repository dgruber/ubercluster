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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgruber/drmaa2"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

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

func showJobsInState(clusteraddress, state string) {
	request := fmt.Sprintf("%s%s%s", clusteraddress, "/monitoring?state=", state)
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
func submitJob(clusteraddress, jobname, cmd, arg, queue string) {
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
	url := fmt.Sprintf("%s%s", clusteraddress, "/session")

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

func showMachinesQueues(clusteraddress, req, filter string) {
	request := fmt.Sprintf("%s/monitoring?%s=%s", clusteraddress, req, filter)
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

// setClusterAddress searches the address of the cluster to contact to
// in the configuration ("default" point to default cluster)
func getClusterAddress(cluster string) string {
	var clusteraddress string
	for i, _ := range config.Cluster {
		if cluster == config.Cluster[i].Name {
			clusteraddress = config.Cluster[i].Address
			clusteraddress = fmt.Sprintf("%s/%s", clusteraddress, config.Cluster[i].ProtocolVersion)
			break
		}
	}
	if clusteraddress == "" {
		fmt.Println("Cluster name %s not found in configuration.", cluster)
		os.Exit(1)
	}
	return clusteraddress
}

var (
	app     = kingpin.New("d2stat", "A tool which can interact with multiple compute clusters.")
	verbose = app.Flag("verbose", "Enables enhanced logging for debugging.").Bool()
	cluster = app.Flag("cluster", "Cluster name to interact with.").Default("default").String()

	show             = app.Command("show", "Displays information about connected clusters.")
	showJob          = show.Command("job", "Information about a particular job.")
	showJobId        = showJob.Arg("id", "Id of job").String()
	showJobByState   = show.Command("jobstate", "All jobs in a specific state (r/p/all).")
	showJobByStateId = showJobByState.Arg("state", "State of jobs to show.").Default("r").String()
	showMachine      = show.Command("machine", "Information about compute hosts.")
	showMachineName  = showMachine.Arg("name", "Name of machine (or \"all\" for all.").Default("all").String()
	showQueue        = show.Command("queue", "Information about queues.")
	showQueueName    = showQueue.Arg("name", "Name of queue to show.").Default("all").String()

	run        = app.Command("run", "Submits an application to a cluster.")
	runCommand = run.Arg("command", "Command to submit.").Required().String()
	runArg     = run.Flag("arg", "Argument of the command.").Default("").String()
	runName    = run.Flag("name", "Reference name of the command.").Default("").String()
	runQueue   = run.Flag("queue", "Queue name in which to submit.").Default("").String()

	cfg     = app.Command("config", "Configuration of cluster proxies.")
	cfgList = cfg.Command("list", "Lists all configured cluster proxies.")
)

func main() {
	p := kingpin.MustParse(app.Parse(os.Args[1:]))

	if *verbose {
		log.SetOutput(os.Stdout)
	}
	// save an config example
	saveDummyConfig()
	// read in configuration
	readConfig()

	// based on cluster name create the address to send requests
	clusteraddress := getClusterAddress(*cluster)
	if p == showJob.FullCommand() {
		fmt.Println("show command selected")
		if showJobId != nil && *showJobId != "" {
			fmt.Println("show job details: ", *showJobId)
			showJobDetails(clusteraddress, *showJobId)
		} else {
			fmt.Println("Job id misssing.")
		}
	}

	if p == cfgList.FullCommand() {
		listConfig(clusteraddress)
	}

	if p == showJobByState.FullCommand() {
		showJobsInState(clusteraddress, *showJobByStateId)
	}

	if p == showMachine.FullCommand() {
		showMachines(clusteraddress, *showMachineName)
	}
	if p == showQueue.FullCommand() {
		showQueues(clusteraddress, *showQueueName)
	}

	if p == run.FullCommand() {
		submitJob(clusteraddress, *runName, *runCommand, *runArg, *runQueue)
	}
}
