/*
   Copyright 2014 Daniel Gruber, Univa, My blog http://www.gridengine.eu

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
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"log"
	"os"
)

// Disable logging by default
func init() {
	log.SetOutput(ioutil.Discard)
}

var (
	app     = kingpin.New("uc", "A tool which can interact with multiple compute clusters.")
	verbose = app.Flag("verbose", "Enables enhanced logging for debugging.").Bool()
	cluster = app.Flag("cluster", "Cluster name to interact with.").Default("default").String()

	show            = app.Command("show", "Displays information about connected clusters.")
	showJob         = show.Command("job", "Information about a particular job.")
	showJobStateId  = showJob.Flag("state", "Show only jobs in that state (r/q/h/s/R/Rh/d/f/u/all).").Default("all").String()
	showJobId       = showJob.Arg("id", "Id of job").Default("").String()
	showJobUser     = showJob.Flag("user", "Shows only jobs of a particular user.").Default("").String()
	showMachine     = show.Command("machine", "Information about compute hosts.")
	showMachineName = showMachine.Arg("name", "Name of machine (or \"all\" for all.").Default("all").String()
	showQueue       = show.Command("queue", "Information about queues.")
	showQueueName   = showQueue.Arg("name", "Name of queue to show.").Default("all").String()

	run         = app.Command("run", "Submits an application to a cluster.")
	runCommand  = run.Arg("command", "Command to submit.").Required().String()
	runArg      = run.Flag("arg", "Argument of the command.").Default("").String()
	runName     = run.Flag("name", "Reference name of the command.").Default("").String()
	runQueue    = run.Flag("queue", "Queue name for the job.").Default("").String()
	runCategory = run.Flag("category", "Job category / job class of the job.").Default("").String()

	// operations on job
	terminate      = app.Command("terminate", "Terminate operation.")
	terminateJob   = terminate.Command("job", "Terminates (ends) a job in a cluster.")
	terminateJobId = terminateJob.Arg("jobid", "Id of the job to terminate.").Default("").String()

	suspend      = app.Command("suspend", "Suspend operation.")
	suspendJob   = suspend.Command("job", "Suspends (pauses) a job in a cluster.")
	suspendJobId = suspendJob.Arg("jobid", "Id of the job to suspend.").Default("").String()

	resume      = app.Command("resume", "Resume operation.")
	resumeJob   = resume.Command("job", "Resumes a suspended job in a cluster.")
	resumeJobId = resumeJob.Arg("jobid", "Id of the job to resume.").Default("").String()

	// configuration
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

	switch p {
	case showJob.FullCommand():
		if showJobId != nil && *showJobId != "" {
			log.Println("showJobId: ", *showJobId)
			showJobDetails(clusteraddress, *showJobId)
		} else {
			showJobs(clusteraddress, *showJobStateId, *showJobUser)
		}
	case cfgList.FullCommand():
		listConfig(clusteraddress)
	case showMachine.FullCommand():
		showMachines(clusteraddress, *showMachineName)
	case showQueue.FullCommand():
		showQueues(clusteraddress, *showQueueName)
	case run.FullCommand():
		submitJob(clusteraddress, *runName, *runCommand, *runArg, *runQueue, *runCategory)
	case terminateJob.FullCommand():
		performOperation(clusteraddress, "ubercluster", "terminate", *terminateJobId)
	case suspendJob.FullCommand():
		performOperation(clusteraddress, "ubercluster", "suspend", *suspendJobId)
	case resumeJob.FullCommand():
		performOperation(clusteraddress, "ubercluster", "resume", *resumeJobId)
	}
}
