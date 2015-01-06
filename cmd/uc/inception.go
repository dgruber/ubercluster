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
)

type inception struct {
	config Config // uc configuration object
}

// Implements the ProxyImplementer interface

func (i *inception) GetJobInfosByFilter(filtered bool, filter ubercluster.JobInfo) []ubercluster.JobInfo {
	return nil
}

func (i *inception) GetJobInfo(jobid string) *ubercluster.JobInfo {
	return nil
}

func (i *inception) GetAllMachines(machines []string) ([]ubercluster.Machine, error) {
	return nil, nil
}

func (i *inception) GetAllQueues(queues []string) ([]ubercluster.Queue, error) {
	return nil, nil
}

func (i *inception) GetAllCategories() ([]string, error) {
	return nil, nil
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
	ubercluster.ProxyListenAndServe(address, "", "", &incept)
}
