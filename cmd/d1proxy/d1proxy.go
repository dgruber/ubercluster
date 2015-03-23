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
	"github.com/dgruber/ubercluster/pkg/proxy"
	"github.com/dgruber/ubercluster/pkg/types"
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
func convertDRMAAJobTemplate(s drmaa.Session, jt types.JobTemplate) (*drmaa.JobTemplate, error) {
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
func (dp *drmaa1Proxy) RunJob(template types.JobTemplate) (jobid string, err error) {
	// file path fix when the app is uploaded
	localFile := template.WorkingDirectory + "/" + template.RemoteCommand
	log.Println("Local file: ", localFile)
	if fi, statErr := os.Stat(localFile); statErr == nil {
		if fi.IsDir() == false {
			// since we have a file in staging area we execute it :/
			log.Println("Adding path to remote command")
			template.RemoteCommand = localFile
		}
	}
	if jt, convErr := convertDRMAAJobTemplate(dp.Session, template); convErr != nil {
		log.Println("Error during job template conversion: ", convErr)
		err = convErr
	} else {
		if id, runErr := dp.Session.RunJob(jt); err != nil {
			err = runErr
		} else {
			jobid = id
		}
	}
	return jobid, err
}

// JobOperation changes the state of a job in the system. Is required by the
// ProxyImplementer interface.
func (dp *drmaa1Proxy) JobOperation(jobsessionname, operation, jobid string) (out string, err error) {
	// in DRMAA1 we irgnore the job session name since we don't have any
	switch operation {
	case "suspend":
		if opErr := dp.Session.SuspendJob(jobid); opErr != nil {
			err = opErr
		} else {
			out = "Suspended Job"
		}
	case "resume":
		if opErr := dp.Session.ResumeJob(jobid); opErr != nil {
			err = opErr
		} else {
			out = "Resumed Job"
		}
	case "terminate":
		if opErr := dp.Session.TerminateJob(jobid); opErr != nil {
			err = opErr
		} else {
			out = "Terminated Job"
		}
		// TODO adding hold and resume
	default:
		log.Println("JobOperation unknown operation ", operation)
		err = errors.New("Unknown operation: " + operation)
	}
	return out, err
}

// GetJobInfosByFilter is not available in DRMAA1. Is required by the
// ProxyImplementer interface.
func (dp *drmaa1Proxy) GetJobInfosByFilter(filtered bool, filter types.JobInfo) []types.JobInfo {
	// filtering is not supported in DRMAA1
	return nil
}

// convertDRMAAState converts a DRMAA state into a DRMAA2 state
func convertDRMAAStateString(ds string) types.JobState {
	switch ds {
	case "Undetermined":
		return types.Undetermined
	case "QueuedActive":
		return types.Queued
	case "SystemHold":
		return types.QueuedHeld
	case "UserHold":
		return types.QueuedHeld
	case "UserSystemHold":
		return types.QueuedHeld
	case "Running":
		return types.Running
	case "SystemSuspended":
		return types.Suspended
	case "UserSuspended":
		return types.Suspended
	case "UserSystemSuspended":
		return types.Suspended
	case "Done":
		return types.Done
	case "Failed":
		return types.Failed
	}
	return types.Undetermined
}

func convertDRMAAState(ds drmaa.PsType) types.JobState {
	switch ds {
	case drmaa.PsUndetermined:
		return types.Undetermined
	case drmaa.PsQueuedActive:
		return types.Queued
	case drmaa.PsSystemOnHold:
		return types.QueuedHeld
	case drmaa.PsUserOnHold:
		return types.QueuedHeld
	case drmaa.PsUserSystemOnHold:
		return types.QueuedHeld
	case drmaa.PsRunning:
		return types.Running
	case drmaa.PsSystemSuspended:
		return types.Suspended
	case drmaa.PsUserSuspended:
		return types.Suspended
	case drmaa.PsUserSystemSuspended:
		return types.Suspended
	case drmaa.PsDone:
		return types.Done
	case drmaa.PsFailed:
		return types.Failed
	}
	return types.Undetermined
}

// GetJobInfo returns information about a job. Probablamtic in
// DRMAA1 since you have to wait for the job being finished.
func (dp *drmaa1Proxy) GetJobInfo(jobid string) *types.JobInfo {
	// assume that jobid is PID
	state, err := dp.Session.JobPs(jobid)
	if err == nil {
		// no more information availble in DRMAA1!
		var ji types.JobInfo
		ji.State = convertDRMAAState(state)
		ji.Id = jobid
		// TODO plugin for vendors to make it more useful?
		return &ji
	} else {
		log.Println(err)
	}
	return nil
}

// GetAllMachines is not available in DRMAA.
func (dp *drmaa1Proxy) GetAllMachines(machines []string) ([]types.Machine, error) {
	// no machines in DRMAA1 -> we need to call the DRM system
	return nil, nil
}

// GetAllQueues is not really helpful since there is no notion of queues
// in DRMAA1.
func (dp *drmaa1Proxy) GetAllQueues(queues []string) ([]types.Queue, error) {
	// no queues in DRMAA1
	var q types.Queue
	q.Name = "DRMAA"
	return []types.Queue{q}, nil
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

	var sc proxy.SecConfig
	sc.OTP = *otp

	proxy.ProxyListenAndServe(*cliPort, *certFile, *keyFile, sc, &d1)
	defer d1.Session.Exit()
}
