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
	"fmt"
	"github.com/dgruber/drmaa2"
	"github.com/dgruber/ubercluster/pkg/types"
	"log"
)

func (d2p *drmaa2proxy) initializeDRMAA2(jsName string) error {
	var sm drmaa2.SessionManager
	var err error

	d2p.sm = sm

	if d2p.ms, err = sm.OpenMonitoringSession(""); err != nil {
		log.Fatal("Couldn't open DRMAA2 MonitoringSession")
	}

	if d2p.js, err = sm.CreateJobSession(jsName, ""); err != nil {
		log.Println("(proxy): Job session ", jsName, " exists already. Reopen it.")
		if d2p.js, err = sm.OpenJobSession(jsName); err != nil {
			log.Fatal("(proxy): Couldn't open job session: ", err)
		}
	}
	return nil
}

// Returns a DRMAA2 JobInfo struct based on the given jobid.
func getJobInfo(ms *drmaa2.MonitoringSession, jobid string) *drmaa2.JobInfo {
	var jobinfo *drmaa2.JobInfo
	filter := drmaa2.CreateJobInfo()
	filter.Id = jobid
	if job, err := ms.GetAllJobs(&filter); err != nil || job == nil {
		filter.Id = fmt.Sprintf("%s.1", jobid)
		if job2, err2 := ms.GetAllJobs(&filter); err2 != nil || job2 == nil {
			log.Printf("No job found")
		} else {
			jobinfo, _ = job2[0].GetJobInfo()
		}
	} else {
		log.Printf("amount of matching jobs %d\n", len(job))
		if len(job) >= 1 {
			jobinfo, _ = job[0].GetJobInfo()
		}
	}
	return jobinfo
}

func getJobInfosByFilter(ms *drmaa2.MonitoringSession, filter *drmaa2.JobInfo) []drmaa2.JobInfo {
	if job, err := ms.GetAllJobs(filter); err != nil || job == nil {
		if err != nil {
			fmt.Printf("Error during GetAllJobs(): %s\n", err)
		} else {
			log.Println("No job in that state found!")
		}
	} else {
		log.Printf("amount of matching jobs %d\n", len(job))
		if len(job) >= 1 {
			ji := make([]drmaa2.JobInfo, 0, 500)
			for i := range job {
				jinfo, _ := job[i].GetJobInfo()
				ji = append(ji, *jinfo)
			}
			return ji
		}
	}
	return nil
}

// getJobInfoByState returns an array of JobInfo objects
// of jobs matching a given job state (or nil)
func getJobInfoByState(ms *drmaa2.MonitoringSession, state string) []drmaa2.JobInfo {
	jobinfo := drmaa2.CreateJobInfo()
	filter := &jobinfo
	switch state {
	case "r":
		filter.State = drmaa2.Running
	case "q":
		filter.State = drmaa2.Queued
	case "h":
		filter.State = drmaa2.QueuedHeld
	case "s":
		filter.State = drmaa2.Suspended
	case "R":
		filter.State = drmaa2.Requeued
	case "Rh":
		filter.State = drmaa2.RequeuedHeld
	case "d":
		filter.State = drmaa2.Done
	case "f":
		filter.State = drmaa2.Failed
	case "u":
		filter.State = drmaa2.Undetermined
	case "all":
		// no filter, we need all jobs
		filter = nil
	default:
		filter.State = drmaa2.Done
	}

	if job, err := ms.GetAllJobs(filter); err != nil || job == nil {
		if err != nil {
			fmt.Printf("Error during GetAllJobs(): %s\n", err)
		} else {
			log.Println("No job in that state found!")
		}
	} else {
		log.Printf("amount of matching jobs %d\n", len(job))
		if len(job) >= 1 {
			ji := make([]drmaa2.JobInfo, 0, 500)
			for i := range job {
				jinfo, _ := job[i].GetJobInfo()
				ji = append(ji, *jinfo)
			}
			return ji
		}
	}
	return nil
}

// Unfortunatley we need to convert here. This is going
// to be removed as soon as there is a pure Go DRMAA2
// interface to work with.

func ConvertD2JobInfo(ji drmaa2.JobInfo) (uc types.JobInfo) {
	uc.Id = ji.Id
	uc.ExitStatus = ji.ExitStatus
	uc.TerminatingSignal = ji.TerminatingSignal
	uc.Annotation = ji.Annotation
	// TODO
	uc.State = (types.JobState)(ji.State)
	uc.SubState = ji.SubState
	uc.AllocatedMachines = make([]string, len(ji.AllocatedMachines), len(ji.AllocatedMachines))
	copy(uc.AllocatedMachines, ji.AllocatedMachines)
	uc.SubmissionMachine = ji.SubmissionMachine
	uc.JobOwner = ji.JobOwner
	uc.Slots = ji.Slots
	uc.QueueName = ji.QueueName
	uc.WallclockTime = ji.WallclockTime
	uc.CPUTime = ji.CPUTime
	uc.SubmissionTime = ji.SubmissionTime
	uc.DispatchTime = ji.DispatchTime
	uc.FinishTime = ji.FinishTime
	return uc
}

func ConvertUCJobInfo(ji types.JobInfo) (uc drmaa2.JobInfo) {
	uc.Id = ji.Id
	uc.ExitStatus = ji.ExitStatus
	uc.TerminatingSignal = ji.TerminatingSignal
	uc.Annotation = ji.Annotation
	uc.State = (drmaa2.JobState)(ji.State)
	uc.SubState = ji.SubState
	uc.AllocatedMachines = make([]string, len(ji.AllocatedMachines), len(ji.AllocatedMachines))
	copy(uc.AllocatedMachines, ji.AllocatedMachines)
	uc.SubmissionMachine = ji.SubmissionMachine
	uc.JobOwner = ji.JobOwner
	uc.Slots = ji.Slots
	uc.QueueName = ji.QueueName
	uc.WallclockTime = ji.WallclockTime
	uc.CPUTime = ji.CPUTime
	uc.SubmissionTime = ji.SubmissionTime
	uc.DispatchTime = ji.DispatchTime
	uc.FinishTime = ji.FinishTime
	return uc
}

func ConvertD2Machine(il []drmaa2.Machine) (ol []types.Machine) {
	ol = make([]types.Machine, 0, 0)
	for _, i := range il {
		var o types.Machine
		o.Name = i.Name
		o.Available = i.Available
		o.Sockets = i.Sockets
		o.CoresPerSocket = i.CoresPerSocket
		o.ThreadsPerCore = i.ThreadsPerCore
		o.Load = i.Load
		o.PhysicalMemory = i.PhysicalMemory
		o.VirtualMemory = i.VirtualMemory
		o.Architecture = (types.CPU)(i.Architecture)
		o.OSVersion = (types.Version)(i.OSVersion)
		o.OS = (types.OS)(i.OS)
		ol = append(ol, o)
	}
	return ol
}

func ConvertD2Queue(il []drmaa2.Queue) (ol []types.Queue) {
	ol = make([]types.Queue, 0, 0)
	for _, i := range il {
		var o types.Queue
		o.Name = i.Name
		ol = append(ol, o)
	}
	return ol
}

func ConvertD2Sessions(il []string) (ol []types.Session) {
	ol = make([]types.Session, 0, 0)
	for _, i := range il {
		var o types.Session
		o.Name = i
		ol = append(ol, o)
	}
	return ol
}

func ConvertUCJobTemplate(u types.JobTemplate) (jt drmaa2.JobTemplate) {
	jt.RemoteCommand = u.RemoteCommand
	jt.Args = make([]string, len(u.Args), len(u.Args))
	copy(jt.Args, u.Args)
	jt.SubmitAsHold = u.SubmitAsHold
	jt.ReRunnable = u.ReRunnable
	jt.JobEnvironment = copyMap(u.JobEnvironment)
	jt.WorkingDirectory = u.WorkingDirectory
	jt.JobCategory = u.JobCategory
	jt.Email = make([]string, len(u.Email), len(u.Email))
	copy(jt.Email, u.Email)
	jt.EmailOnStarted = u.EmailOnStarted
	jt.EmailOnTerminated = u.EmailOnTerminated
	jt.JobName = u.JobName
	jt.InputPath = u.InputPath
	jt.OutputPath = u.OutputPath
	jt.ErrorPath = u.ErrorPath
	jt.JoinFiles = u.JoinFiles
	jt.ReservationId = u.ReservationId
	jt.QueueName = u.QueueName
	jt.MaxSlots = u.MaxSlots
	jt.MinSlots = u.MinSlots
	jt.Priority = u.Priority
	jt.CandidateMachines = make([]string, len(u.CandidateMachines), len(u.CandidateMachines))
	copy(jt.CandidateMachines, u.CandidateMachines)
	jt.MinPhysMemory = u.MinPhysMemory
	jt.MachineOs = u.MachineOs
	jt.MachineArch = u.MachineArch
	jt.StartTime = u.StartTime
	jt.DeadlineTime = u.DeadlineTime
	jt.StageInFiles = copyMap(u.StageInFiles)
	jt.StageOutFiles = copyMap(u.StageOutFiles)
	jt.ResourceLimits = copyMap(u.ResourceLimits)
	jt.AccountingId = u.AccountingId
	return jt
}

func copyMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
