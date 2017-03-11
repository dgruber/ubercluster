/*
   Copyright 2014 Daniel Gruber, Univa, My blog: http://www.gridengine.eu

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
	"github.com/dgruber/ubercluster/pkg/http_helper"
	"github.com/dgruber/ubercluster/pkg/output"
	"github.com/dgruber/ubercluster/pkg/proxy"
	"github.com/dgruber/ubercluster/pkg/types"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"os"
)

func getJob(clusteraddress, jobid string) (types.JobInfo, error) {
	request := fmt.Sprintf("%s%s%s", clusteraddress, "/msession/jobinfo/", jobid)
	log.Println("Requesting:" + request)
	resp, err := http_helper.UberGet(*otp, request)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var jobinfo types.JobInfo
	if err := decoder.Decode(&jobinfo); err != nil {
		return jobinfo, err
	}
	return jobinfo, nil
}

func showJobDetails(clustername, jobid string, of output.OutputFormater) {
	jobinfo, err := getJob(clustername, jobid)
	if err == nil {
		of.PrintJobDetails(jobinfo)
	} else {
		fmt.Println("Error: ", err)
	}
}

func getJobs(clusteraddress, state, user string) []types.JobInfo {
	firstSet := false
	request := fmt.Sprintf("%s%s", clusteraddress, "/msession/jobinfos")
	if state != "" && state != "all" {
		firstSet = true
		request = fmt.Sprintf("%s%s%s", request, "?state=", state)
	}
	if user != "" {
		if firstSet == true {
			request = fmt.Sprintf("%s%s", request, "&")
		} else {
			request = fmt.Sprintf("%s%s", request, "?")
		}
		request = fmt.Sprintf("%s%s%s", request, "user=", user)
	}
	log.Println("Requesting:" + request)
	resp, err := http_helper.UberGet(*otp, request)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var joblist []types.JobInfo
	decoder.Decode(&joblist)
	log.Println(joblist)

	return joblist
}

func showJobs(clusteraddress, state, user string, of output.OutputFormater) {
	joblist := getJobs(clusteraddress, state, user)
	for index := range joblist {
		of.PrintJobDetails(joblist[index])
		fmt.Println()
	}
	if len(joblist) == 0 {
		if state != "all" {
			fmt.Printf("No job in state %s found.\n", state)
		} else {
			fmt.Printf("No job found.\n")
		}
	}
}

// getYubiKey requests a one time password on command line from
// the user and returns it - the one time password is created
// by pressing the yubikey button
func getYubiKey() string {
	fmt.Printf("Press yubikey button: ")
	if pw, err := terminal.ReadPassword(0); err != nil {
		fmt.Println("Error reading in password from stdin: ", err)
		os.Exit(1)
	} else {
		return string(pw)
	}
	return ""
}

// submitJob creates a new job in the given cluster
func submitJob(clusteraddress, clustername, jobname, cmd, arg, queue, category, otp string) {
	var jt types.JobTemplate
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
	if resp, err := http_helper.UberPost(otp, url, "application/json",
		bytes.NewBuffer(jtb)); err != nil {
		fmt.Println("Job submission error: ", err)
	} else {
		// fmt.Println("Job submitted successfully: ", resp.Status)
		decoder := json.NewDecoder(resp.Body)
		var answer proxy.RunJobResult
		if err := decoder.Decode(&answer); err != nil {
			fmt.Println("Error during decoding: ", err)
		} else {
			fmt.Println("Jobid: ", answer.JobId)
			fmt.Println("Cluster: ", clustername)
		}
	}
}

func showQueues(clustername, queue string, of output.OutputFormater) {
	showMachinesQueues(clustername, "queues", queue, of)
}

func showMachines(clustername, machine string, of output.OutputFormater) {
	showMachinesQueues(clustername, "machines", machine, of)
}

func createRequestMachinesQueues(clusteraddress, req, filter string) string {
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
	return request
}

func getQueues(clusteraddress, filter string) ([]types.Queue, error) {
	resp, err := http_helper.UberGet(*otp, createRequestMachinesQueues(clusteraddress, "queues", filter))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var queuelist []types.Queue
	if err := decoder.Decode(&queuelist); err != nil {
		fmt.Println("Error during decoding: ", err)
		return nil, err
	}
	return queuelist, nil
}

func getMachines(clusteraddress, filter string) ([]types.Machine, error) {
	resp, err := http_helper.UberGet(*otp, createRequestMachinesQueues(clusteraddress, "machines", filter))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var machinelist []types.Machine
	if err := decoder.Decode(&machinelist); err != nil {
		fmt.Println("Error during decoding: ", err)
		return nil, err
	}
	return machinelist, nil
}

func showMachinesQueues(clusteraddress, req, filter string, of output.OutputFormater) {
	log.Println("showMachineQueues: ", clusteraddress, req, filter)
	if req == "machines" {
		if machinelist, err := getMachines(clusteraddress, filter); err == nil {
			for index := range machinelist {
				//emulateQhost(machinelist[index])
				of.PrintMachine(machinelist[index])
			}
		}
	} else if req == "queues" {
		if queuelist, err := getQueues(clusteraddress, filter); err == nil {
			log.Println("Queuelist: ", queuelist)
			for index := range queuelist {
				fmt.Println(queuelist[index].Name)
				// TODO
			}
		}
	}
}

// performOperation sends request to perform an operation on a particular
// job to a connected cluster (to its proxy).
// The request url is: jsession/<jobsessionname>/<operation>/jobnumber
func performOperation(clusteraddress, jsession, operation, jobId string) {
	url := fmt.Sprintf("%s/jsession/%s/%s/%s", clusteraddress, jsession, operation, jobId)
	log.Println("Requesting:" + url)
	buffer := bytes.NewBuffer([]byte(""))
	if resp, err := http_helper.UberPost(*otp, url, "application/json", buffer); err != nil {
		fmt.Println("Error during post: ", err)
	} else {
		log.Println("Status of request:", resp.Status)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
	}
}

func getJobCategories(clusteraddress, jsession, category string) []string {
	var url string
	if category == "all" || category == "" {
		url = fmt.Sprintf("%s/jsession/%s/jobcategories", clusteraddress, jsession)
	} else {
		url = fmt.Sprintf("%s/jsession/%s/jobcategory/%s", clusteraddress, jsession, category)
	}
	log.Println("Requesting:" + url)
	if resp, err := http_helper.UberGet(*otp, url); err != nil {
		log.Fatal(err)
		os.Exit(1)
	} else {
		defer resp.Body.Close()
		if category == "all" || category == "" {
			var catList []string
			json.NewDecoder(resp.Body).Decode(&catList)
			return catList
		} else {
			var cat string
			json.NewDecoder(resp.Body).Decode(&cat)
			return []string{cat}
		}
	}
	return nil
}

func showJobCategories(clusteraddress, jsession, category string) {
	for _, cat := range getJobCategories(clusteraddress, jsession, category) {
		fmt.Println(cat)
	}
}

func getJobSessions(clusteraddress, jsession string) []string {
	url := fmt.Sprintf("%s/jsessions", clusteraddress)
	log.Println("Requesting:" + url)
	if resp, err := http_helper.UberGet(*otp, url); err != nil {
		log.Fatal(err)
		os.Exit(1)
	} else {
		defer resp.Body.Close()
		var jsList []string
		json.NewDecoder(resp.Body).Decode(&jsList)
		found := false
		if jsession != "all" {
			for _, js := range jsList {
				if js == jsession {
					found = true
				}
			}
			if found == true {
				return []string{jsession}
			} else {
				return []string{}
			}
		}
		return jsList
	}
	return nil
}

// showJobSessions requests all job sessions available on the
// given cluster and prints them out to the user.
func showJobSessions(clusteraddress, jsession string) {
	jSessions := getJobSessions(clusteraddress, jsession)
	if len(jSessions) >= 1 {
		for _, js := range jSessions {
			fmt.Println(js)
		}
	} else {
		if jsession == "all" {
			fmt.Println("No job session found.")
			os.Exit(1)
		}
		fmt.Printf("Job session %s does not exist.\n", jsession)
		os.Exit(1)
	}
}
