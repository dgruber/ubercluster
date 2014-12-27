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
	"github.com/dgruber/ubercluster"
	"log"
	"net/http"
	"os"
)

func showJobDetails(clustername, jobid string) {
	request := fmt.Sprintf("%s%s%s", clustername, "/msession/jobinfo/", jobid)
	log.Println("Requesting:" + request)
	resp, err := http.Get(request)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var jobinfo ubercluster.JobInfo
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
	var joblist []ubercluster.JobInfo
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
func submitJob(clusteraddress, jobname, cmd, arg, queue, category string) {
	var jt ubercluster.JobTemplate
	// fill a DRMAA2 job template and send it over to the proxy
	jt.RemoteCommand = cmd
	jt.JobName = jobname
	if arg != "" {
		jt.Args = []string{arg}
	}
	jt.QueueName = queue
	if category != "" {
		jt.JobCategory = category
	}
	jtb, _ := json.Marshal(jt)

	// create URL of cluster to send the job to
	url := fmt.Sprintf("%s%s", clusteraddress, "/jsession/default/run")
	log.Println("POST to URL:", url)
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
	var request string

	if filter == "all" {
		request = fmt.Sprintf("%s/msession/%s", clusteraddress, req)
	} else {
		// filter for a specific queue or machine
		if req == "machines" {
			request = fmt.Sprintf("%s/msession/machine/%s", clusteraddress, filter)
		} else {
			request = fmt.Sprintf("%s/msession/queue/%s", clusteraddress, filter)
		}
	}
	log.Println("Requesting:" + request)
	resp, err := http.Get(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	if req == "machines" {
		var machinelist []ubercluster.Machine
		if err := decoder.Decode(&machinelist); err != nil {
			fmt.Println("Error during decoding: ", err)
			os.Exit(1)
		}
		for index, _ := range machinelist {
			emulateQhost(machinelist[index])
		}
	} else if req == "queues" {
		var queuelist []ubercluster.Queue
		if err := decoder.Decode(&queuelist); err != nil {
			fmt.Println("Error during decoding: ", err)
			os.Exit(1)
		}
		for index, _ := range queuelist {
			fmt.Println(queuelist[index].Name)
		}
	}
}
