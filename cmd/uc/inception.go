/*
   Copyright 2015 Daniel Gruber, Univa, My blog: http://www.gridengine.eu

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

// Run uc as proxy itself. Allows to stack clusters of cluster recursively.

import (
	"fmt"
	"github.com/dgruber/ubercluster"
	"log"
)

type inception struct {
	inceptionAddress string // address of uc itself
	config           Config // uc configuration object
}

// Implements the ProxyImplementer interface

func (i *inception) GetJobInfosByFilter(filtered bool, filter ubercluster.JobInfo) []ubercluster.JobInfo {
	return nil
}

func (i *inception) GetJobInfo(jobid string) *ubercluster.JobInfo {
	// search job id in all connected clusters
	// if it has a postfix - only in that cluster
	// 1301@mybiggridenginecluster search 1301 in the given cluster

	return nil
}

func (i *inception) GetAllMachines(machines []string) ([]ubercluster.Machine, error) {
	allmachines := make([]ubercluster.Machine, 0, 0)
	for _, c := range i.config.Cluster {
		log.Println("Requesting from: ", c.Address)
		// we don't request our own address...
		if addr := fmt.Sprintf("%s/", c.Address); addr == i.inceptionAddress {
			continue
		}
		if ms, err := getMachines(getClusterAddress(c.Name), "all"); err == nil {
			allmachines = append(allmachines, ms...)
			log.Println("Appending: ", allmachines)
		} else {
			log.Println("Error while requesting machines from ", c.Name, err)
		}
		// TODO filter according request
		// TODO remove duplicates
	}
	return allmachines, nil
}

// GetAllQueues returns all queue names from all clusters which are
// connected to the uc tool.
func (i *inception) GetAllQueues(queues []string) ([]ubercluster.Queue, error) {
	allqueues := make([]ubercluster.Queue, 0, 0)
	// TODO go functions of course
	for _, c := range i.config.Cluster {
		log.Println("Requesting from: ", c.Address)
		// we don't request our own address...
		if addr := fmt.Sprintf("%s/", c.Address); addr == i.inceptionAddress {
			continue
		}
		if qs, err := getQueues(getClusterAddress(c.Name), "all"); err == nil {
			allqueues = append(allqueues, qs...)
			log.Println("Appending: ", allqueues)
		} else {
			log.Println("Error while requesting queues from ", c.Name, err)
		}
		// TODO filter according request
		// TODO remove duplicates
	}
	return allqueues, nil
}

func (i *inception) GetAllCategories() ([]string, error) {
	cat := make([]string, 0, 0)
	for _, c := range i.config.Cluster {
		log.Println("Requesting from: ", c.Address)
		if addr := fmt.Sprintf("%s/", c.Address); addr == i.inceptionAddress {
			log.Println("Skipping own address")
			continue
		}
		cat = append(cat, getJobCategories(getClusterAddress(c.Name), "ubercluster", "all")...)
	}
	return cat, nil
}

func (i *inception) DRMSVersion() string {
	return "0.1"
}

func (i *inception) DRMSName() string {
	return "ubercluster"
}

func (i *inception) RunJob(template ubercluster.JobTemplate) (string, error) {
	return "", nil
}

func (i *inception) JobOperation(jobsessionname, operation, jobid string) (string, error) {
	return "", nil
}

// start uc as proxy
func inceptionMode(address string) {
	var incept inception
	incept.config = config // configuration contains all connected clusters
	fmt.Println("Starting uc in inception mode as proxy listing at address: ", address)
	ubercluster.ProxyListenAndServe(address, "", "", "", &incept)
}
