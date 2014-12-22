package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgruber/drmaa2"
	"log"
	"net/http"
	"os"
)

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
