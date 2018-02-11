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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/dgruber/ubercluster/pkg/http_helper"
	"github.com/dgruber/ubercluster/pkg/output"
	"github.com/dgruber/ubercluster/pkg/proxy"
	"github.com/dgruber/ubercluster/pkg/types"

	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Request struct {
	otp    *string
	client *http.Client
}

func NewRequest(certFile string, keyFile string, oneTimePassword *string) *Request {
	var config tls.Config

	if certFile != "" && keyFile != "" {
		fmt.Println("Using certificates")

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Panicln(err.Error())
		}

		certs := []tls.Certificate{cert}

		clientCACert, err := ioutil.ReadFile(certFile)
		if err != nil {
			log.Fatal("Unable to open cert", err)
		}

		clientCertPool := x509.NewCertPool()
		clientCertPool.AppendCertsFromPEM(clientCACert)

		config = tls.Config{
			InsecureSkipVerify: true,
			Certificates:       certs,
			RootCAs:            clientCertPool,
		}
		config.BuildNameToCertificate()
	} else {
		fmt.Println("unsecure client")
		config = tls.Config{
			InsecureSkipVerify: true,
		}
	}

	config.BuildNameToCertificate()

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
		TLSClientConfig:    &config,
	}

	client := &http.Client{Transport: tr}

	return &Request{
		otp:    oneTimePassword,
		client: client,
	}
}

func (r *Request) SelectClusterAddress(cluster, alg string) (string, string, error) {
	// a cluster selection algorithm chooses the right cluster
	switch alg {
	case "rand": // random scheduling
		return GetClusterAddress(MakeNewScheduler(RandomSchedulerType, config, r.client).Impl.SelectCluster())
	case "prob": // probabilistic scheduling
		return GetClusterAddress(MakeNewScheduler(ProbabilisticSchedulerType, config, r.client).Impl.SelectCluster())
	case "load": // load based scheduling
		return GetClusterAddress(MakeNewScheduler(LoadBasedSchedulerType, config, r.client).Impl.SelectCluster())
	}
	if alg != "" {
		fmt.Println("Unkown scheduler selection algorithm: ", alg)
		os.Exit(2)
	}
	return GetClusterAddress(cluster)
}

func (r *Request) GetJob(clusteraddress, jobid string) (types.JobInfo, error) {
	request := fmt.Sprintf("%s%s%s", clusteraddress, "/msession/jobinfo/", jobid)
	log.Println("Requesting:" + request)

	resp, err := http_helper.UberGet(r.client, *otp, request)
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

func (r *Request) ShowJobDetails(clustername, jobid string, of output.OutputFormater) {
	jobinfo, err := r.GetJob(clustername, jobid)
	if err == nil {
		of.PrintJobDetails(jobinfo)
	} else {
		fmt.Println("Error: ", err)
	}
}

func (r *Request) GetJobs(clusteraddress, state, user string) []types.JobInfo {
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
	resp, err := http_helper.UberGet(r.client, *otp, request)
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

func (r *Request) ShowJobs(clusteraddress, state, user string, of output.OutputFormater) {
	joblist := r.GetJobs(clusteraddress, state, user)
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

func (r *Request) RunLocalRequest(otp, clusteraddress, cmd, arg string) {
	url := fmt.Sprintf("%s%s", clusteraddress, "/local/run")
	log.Println("POST to URL:", url)
	rlr := types.RunLocalRequest{
		Command: cmd,
		Arg:     arg,
	}
	body, _ := json.Marshal(rlr)
	resp, err := http_helper.UberPost(r.client, otp, url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Run local error: ", err)
		return
	}
	defer resp.Body.Close()

	var answer string
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error during reading answer from proxy: %s\n", err.Error())
		return
	}
	json.Unmarshal(respBody, &answer)
	fmt.Printf("%s\n", answer)
}

func (r *Request) CreateJobRequest(jobname, cmd, arg, queue, category string) []byte {
	jt := types.JobTemplate{
		RemoteCommand: cmd,
		JobName:       jobname,
		QueueName:     queue,
		JobCategory:   category,
	}
	if arg != "" {
		jt.Args = []string{arg}
	}
	jtb, _ := json.Marshal(jt)
	return jtb
}

// SubmitJob creates a new job in the given cluster
func (r *Request) SubmitJob(clusteraddress, clustername, jobname, cmd, arg, queue, category, otp string) {
	jtb := r.CreateJobRequest(jobname, cmd, arg, queue, category)

	// create URL of cluster to send the job to
	url := fmt.Sprintf("%s%s", clusteraddress, "/jsession/default/run")
	log.Println("POST to URL:", url)
	log.Println("Submit template: ", string(jtb))

	resp, err := http_helper.UberPost(r.client, otp, url, "application/json", bytes.NewBuffer(jtb))
	if err != nil {
		fmt.Printf("Job submission error: %s\n", err.Error())
		return
	}
	defer resp.Body.Close()

	// fmt.Println("Job submitted successfully: ", resp.Status)
	var answer proxy.RunJobResult
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error during reading answer from proxy: %s\n", err.Error())
		return
	}

	err = json.Unmarshal(body, &answer)
	if err != nil {
		fmt.Printf("Error during decoding answer from POSTING to proxy during job submission: %s\n", string(body))
	} else {
		fmt.Println("Job ID: ", answer.JobId)
		fmt.Println("Cluster: ", clustername)
	}
}

func (r *Request) ShowQueues(clustername, queue string, of output.OutputFormater) {
	r.ShowMachinesQueues(clustername, "queues", queue, of)
}

func (r *Request) ShowMachines(clustername, machine string, of output.OutputFormater) {
	r.ShowMachinesQueues(clustername, "machines", machine, of)
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

func (r *Request) GetQueues(clusteraddress, filter string) ([]types.Queue, error) {
	resp, err := http_helper.UberGet(r.client, *otp, createRequestMachinesQueues(clusteraddress, "queues", filter))
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

func (r *Request) GetMachines(clusteraddress, filter string) ([]types.Machine, error) {
	resp, err := http_helper.UberGet(r.client, *otp, createRequestMachinesQueues(clusteraddress, "machines", filter))
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

func (r *Request) ShowMachinesQueues(clusteraddress, req, filter string, of output.OutputFormater) {
	log.Println("showMachineQueues: ", clusteraddress, req, filter)
	if req == "machines" {
		if machinelist, err := r.GetMachines(clusteraddress, filter); err == nil {
			for index := range machinelist {
				//emulateQhost(machinelist[index])
				of.PrintMachine(machinelist[index])
			}
		}
	} else if req == "queues" {
		if queuelist, err := r.GetQueues(clusteraddress, filter); err == nil {
			log.Println("Queuelist: ", queuelist)
			for index := range queuelist {
				fmt.Println(queuelist[index].Name)
				// TODO
			}
		}
	}
}

// PerformOperation sends request to perform an operation on a particular
// job to a connected cluster (to its proxy).
// The request url is: jsession/<jobsessionname>/<operation>/jobnumber
func (r *Request) PerformOperation(clusteraddress, jsession, operation, jobId string) {
	url := fmt.Sprintf("%s/jsession/%s/%s/%s", clusteraddress, jsession, operation, jobId)
	log.Println("Requesting:" + url)
	buffer := bytes.NewBuffer([]byte(""))
	if resp, err := http_helper.UberPost(r.client, *otp, url, "application/json", buffer); err != nil {
		fmt.Println("Error during post: ", err)
	} else {
		log.Println("Status of request:", resp.Status)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
	}
}

func (r *Request) GetJobCategories(clusteraddress, jsession, category string) []string {
	var url string
	if category == "all" || category == "" {
		url = fmt.Sprintf("%s/jsession/%s/jobcategories", clusteraddress, jsession)
	} else {
		url = fmt.Sprintf("%s/jsession/%s/jobcategory/%s", clusteraddress, jsession, category)
	}
	log.Println("Requesting:" + url)
	if resp, err := http_helper.UberGet(r.client, *otp, url); err != nil {
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

func (r *Request) ShowJobCategories(clusteraddress, jsession, category string) {
	for _, cat := range r.GetJobCategories(clusteraddress, jsession, category) {
		fmt.Println(cat)
	}
}

func (r *Request) GetJobSessions(clusteraddress, jsession string) []string {
	url := fmt.Sprintf("%s/jsessions", clusteraddress)
	log.Println("Requesting:" + url)
	if resp, err := http_helper.UberGet(r.client, *otp, url); err != nil {
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

// ShowJobSessions requests all job sessions available on the
// given cluster and prints them out to the user.
func (r *Request) ShowJobSessions(clusteraddress, jsession string) {
	jSessions := r.GetJobSessions(clusteraddress, jsession)
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
