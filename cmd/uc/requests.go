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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgruber/ubercluster"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// uberGet makes an http GET request. Depending on the uc
// configuration (currently cli param) it adds a one time
// password.
func uberGet(request string) (resp *http.Response, err error) {
	newRequest := request
	if *otp != "" {
		// adding http secret key (OTP)
		if strings.Contains(request, "?") {
			newRequest = fmt.Sprintf("%s&otp=%s", request, *otp)
		} else {
			newRequest = fmt.Sprintf("%s?otp=%s", request, *otp)
		}
	}
	log.Println("New request: ", newRequest)
	return http.Get(newRequest)
}

// uberPost is a http.Post replacement which adds otp requests
// and possibly others depending on the configuration.
func uberPost(url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	newUrl := url
	if *otp != "" {
		// adding http secret key (OTP)
		if strings.Contains(url, "?") {
			newUrl = fmt.Sprintf("%s&otp=%s", url, *otp)
		} else {
			newUrl = fmt.Sprintf("%s?otp=%s", url, *otp)
		}
	}
	log.Println("New POST: ", newUrl)
	return http.Post(newUrl, bodyType, body)
}

func showJobDetails(clustername, jobid string) {
	request := fmt.Sprintf("%s%s%s", clustername, "/msession/jobinfo/", jobid)
	log.Println("Requesting:" + request)
	resp, err := uberGet(request)
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

func showJobs(clusteraddress, state, user string) {
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
	resp, err := uberGet(request)
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

func getYubiKey() string {
	fmt.Printf("Press yubikey button: ")
	bio := bufio.NewReader(os.Stdin)
	line, _, _ := bio.ReadLine()
	return string(line)
}

// submitJob creates a new job in the given cluster
func submitJob(clusteraddress, jobname, cmd, arg, queue, category, otp string) {
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
	if otp != "" {
		// it is yubi
		key := getYubiKey()
		url = fmt.Sprintf("%s?otp=", key)
	}
	log.Println("POST to URL:", url)
	log.Println("Submit template: ", string(jtb))
	if resp, err := uberPost(url, "application/json", bytes.NewBuffer(jtb)); err != nil {
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

func showMachinesQueues(clusteraddress, req, filter string) {
	resp, err := uberGet(createRequestMachinesQueues(clusteraddress, req, filter))
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

// performOperation sends request to perform an operation on a particular
// job to a connected cluster (to its proxy).
// The request url is: jsession/<jobsessionname>/<operation>/jobnumber
func performOperation(clusteraddress, jsession, operation, jobId string) {
	url := fmt.Sprintf("%s/jsession/%s/%s/%s", clusteraddress, jsession, operation, jobId)
	log.Println("Requesting:" + url)
	buffer := bytes.NewBuffer([]byte(""))
	if resp, err := uberPost(url, "application/json", buffer); err != nil {
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
	if category == "all" {
		url = fmt.Sprintf("%s/jsession/%s/jobcategories", clusteraddress, jsession)
	} else {
		url = fmt.Sprintf("%s/jsession/%s/jobcategory/%s", clusteraddress, jsession, category)
	}
	log.Println("Requesting:" + url)
	if resp, err := uberGet(url); err != nil {
		log.Fatal(err)
		os.Exit(1)
	} else {
		defer resp.Body.Close()
		if category == "all" {
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
