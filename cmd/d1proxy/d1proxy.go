/*
   Copyright 2015 Daniel Gruber, Univa, Blog: www.gridengine.eu

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

// This is a simple proxy for ubercluster which supports all DRAMA1
// compatible DRM systems (Grid Engine, etc.).
package main

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa"
	"github.com/dgruber/ubercluster"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"log"
	"os"
)

var verbose bool = false

func init() {
	if verbose == false {
		log.SetOutput(ioutil.Discard)
	}
}

// Standard set of CLI parameters.
var (
	app        = kingpin.New("d1proxy", "A proxy server for DRMAA1 compatible cluster schedulers (like Univa Grid Engine).")
	cliVerbose = app.Flag("verbose", "Enables enhanced logging for debugging.").Bool()
	cliPort    = app.Flag("port", "Sets address and port on which proxy is listening.").Default(":8888").String()
	certFile   = app.Flag("certFile", "Path to certification file for secure connections (TLS).").Default("").String()
	keyFile    = app.Flag("keyFile", "Path to key file for secure connections (TLS).").Default("").String()
	otp        = app.Flag("otp", "One time password settings (\"yubikey\") or a fixed shared secret.").Default("").String()
)

// drmaa1Proxy is our internal DRMAA1 DRMS implementation.
type drmaa1Proxy struct {
	Session drmaa.Session
}

// convertDRMAAJobTemplate transforms a DRMAA2 job template (from ubercluster package)
// into a DRMAA1 job template which is used for executing the job
func convertDRMAAJobTemplate(s drmaa.Session, jt ubercluster.JobTemplate) (*drmaa.JobTemplate, error) {
	djt, err := s.AllocateJobTemplate()
	if err != nil {
		log.Println("Error during job template allocation: ", err)
		return nil, err
	}
	if err := djt.SetJobName(jt.JobName); err != nil {
		log.Println("Error during SetJobName: ", err)
	}
	if err := djt.SetRemoteCommand(jt.RemoteCommand); err != nil {
		log.Println("Error during SetRemoteCommand: ", err)
	}
	if err := djt.SetArgs(jt.Args); err != nil {
		log.Println("Error during SetArgs: ", err)
	}
	// TODO we have more parameters
	return &djt, nil
}

// RunJob runs a DRMAA job in the cluster. Is required in order to fulfill
// the  ProxyImplementer interface.
func (dp *drmaa1Proxy) RunJob(template ubercluster.JobTemplate) (string, error) {
	// file path fix when the app is uploaded
	localFile := template.WorkingDirectory + "/" + template.RemoteCommand
	log.Println("Local file: ", localFile)
	if fi, err := os.Stat(localFile); err == nil {
		if fi.IsDir() == false {
			// since we have a file in staging area we execute it :/
			log.Println("Adding path to remote command")
			template.RemoteCommand = localFile
		}
	}
	if jt, err := convertDRMAAJobTemplate(dp.Session, template); err != nil {
		log.Println("Error during job template conversion: ", err)
		return "", err
	} else {
		if jobid, err := dp.Session.RunJob(jt); err != nil {
			return "", err
		} else {
			return jobid, nil
		}
	}
	// unreachable
	return "NOOP", nil
}

// JobOperation changes the state of a job in the system. Is required by the
// ProxyImplementer interface.
func (dp *drmaa1Proxy) JobOperation(jobsessionname, operation, jobid string) (string, error) {
	// in DRMAA1 we irgnore the job session name since we don't have any
	switch operation {
	case "suspend":
		if err := dp.Session.SuspendJob(jobid); err != nil {
			return "", err
		}
		return "Suspended job", nil
	case "resume":
		if err := dp.Session.ResumeJob(jobid); err != nil {
			return "", err
		}
		return "Resumed job", nil
	case "terminate":
		if err := dp.Session.TerminateJob(jobid); err != nil {
			return "", err
		}
		return "Terminated job", nil
		// TODO adding hold and resume
	default:
		log.Println("JobOperation unknown operation ", operation)
		return "", errors.New("Unknown operation: " + operation)
	}
	// unreachable - go ...
	return "NOOP", nil
}

// GetJobInfosByFilter is not available in DRMAA1. Is required by the
// ProxyImplementer interface.
func (dp *drmaa1Proxy) GetJobInfosByFilter(filtered bool, filter ubercluster.JobInfo) []ubercluster.JobInfo {
	// filtering is not supported in DRMAA1
	return nil
}

// convertDRMAAState converts a DRMAA state into a DRMAA2 state
func convertDRMAAStateString(ds string) ubercluster.JobState {
	switch ds {
	case "Undetermined":
		return ubercluster.Undetermined
	case "QueuedActive":
		return ubercluster.Queued
	case "SystemHold":
		return ubercluster.QueuedHeld
	case "UserHold":
		return ubercluster.QueuedHeld
	case "UserSystemHold":
		return ubercluster.QueuedHeld
	case "Running":
		return ubercluster.Running
	case "SystemSuspended":
		return ubercluster.Suspended
	case "UserSuspended":
		return ubercluster.Suspended
	case "UserSystemSuspended":
		return ubercluster.Suspended
	case "Done":
		return ubercluster.Done
	case "Failed":
		return ubercluster.Failed
	}
	return ubercluster.Undetermined
}

func convertDRMAAState(ds drmaa.PsType) ubercluster.JobState {
	switch ds {
	case drmaa.PsUndetermined:
		return ubercluster.Undetermined
	case drmaa.PsQueuedActive:
		return ubercluster.Queued
	case drmaa.PsSystemOnHold:
		return ubercluster.QueuedHeld
	case drmaa.PsUserOnHold:
		return ubercluster.QueuedHeld
	case drmaa.PsUserSystemOnHold:
		return ubercluster.QueuedHeld
	case drmaa.PsRunning:
		return ubercluster.Running
	case drmaa.PsSystemSuspended:
		return ubercluster.Suspended
	case drmaa.PsUserSuspended:
		return ubercluster.Suspended
	case drmaa.PsUserSystemSuspended:
		return ubercluster.Suspended
	case drmaa.PsDone:
		return ubercluster.Done
	case drmaa.PsFailed:
		return ubercluster.Failed
	}
	return ubercluster.Undetermined
}

// GetJobInfo returns information about a job. Probablamtic in
// DRMAA1 since you have to wait for the job being finished.
func (dp *drmaa1Proxy) GetJobInfo(jobid string) *ubercluster.JobInfo {
	// assume that jobid is PID
	if state, err := dp.Session.JobPs(jobid); err != nil {
		log.Println(err)
		return nil
	} else {
		// no more information availble in DRMAA1!
		var ji ubercluster.JobInfo
		ji.State = convertDRMAAState(state)
		ji.Id = jobid
		// TODO plugin for vendors to make it more useful?
		return &ji
	}
	// unreachable - go ...
	return nil
}

// GetAllMachines is not available in DRMAA.
func (dp *drmaa1Proxy) GetAllMachines(machines []string) ([]ubercluster.Machine, error) {
	// no machines in DRMAA1 -> we need to call the DRM system
	return nil, nil
}

// GetAllQueues is not really helpful since there is no notion of queues
// in DRMAA1.
func (dp *drmaa1Proxy) GetAllQueues(queues []string) ([]ubercluster.Queue, error) {
	// no queues in DRMAA1
	var q ubercluster.Queue
	q.Name = "DRMAA"
	return []ubercluster.Queue{q}, nil
}

// GetAllSessions just returns one session name since there is just one
// (unnamed) session possible in DRMAA1.
func (dp *drmaa1Proxy) GetAllSessions(session []string) ([]string, error) {
	// only one session in DRMAA1
	allsessions := make([]string, 0, 0)
	allsessions = append(allsessions, "DRMAA")
	return allsessions, nil
}

// GetAllCategories returns nothing since there is no category listing
// available in DRMAA1.
func (dp *drmaa1Proxy) GetAllCategories() ([]string, error) {
	// no real catgegories in DRMAA1
	return nil, nil
}

// DRMSVersion returns the version of the DRMAA implementation.
func (dp *drmaa1Proxy) DRMSVersion() string {
	return dp.Session.GetDrmaaImplementation()
}

// DRMSName returns the name of the DRM (like Univa Grid Engine or Sun Grid Engine)
func (dp *drmaa1Proxy) DRMSName() string {
	sys, _ := dp.Session.GetDrmSystem()
	return sys
}

func (lp *drmaa1Proxy) DRMSLoad() float64 {
	// TODO some ratio about pending / running jobs respecting the
	// throughput
	return 0.5
}

// InitDRMAA opens a DRMAA session which is going to be used
// by the callbacks.
func InitDRMAA() (drmaa1Proxy, error) {
	var d1p drmaa1Proxy
	s, err := drmaa.MakeSession()
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	} else {
		// add session
		d1p.Session = s
	}
	return d1p, nil
}

func main() {
	var err error

	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *cliVerbose {
		log.SetOutput(os.Stdout)
	}

	d1, err := InitDRMAA()
	if err != nil {
		fmt.Println("Error during initialization of DRMAA: ", err)
		os.Exit(1)
	}

	var sc ubercluster.SecConfig
	sc.OTP = *otp

	ubercluster.ProxyListenAndServe(*cliPort, *certFile, *keyFile, sc, &d1)
	defer d1.Session.Exit()
}
