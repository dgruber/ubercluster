package main

import (
	"fmt"
	"github.com/dgruber/drmaa2"
	"log"
)

// DRMAA2 Sessions required for accessing the cluster
var ms *drmaa2.MonitoringSession
var js *drmaa2.JobSession

func initializeDRMAA2(jsName string) error {
	var sm drmaa2.SessionManager
	var err error

	if ms, err = sm.OpenMonitoringSession(""); err != nil {
		log.Fatal("Couldn't open DRMAA2 MonitoringSession")
	}

	if js, err = sm.CreateJobSession(jsName, ""); err != nil {
		log.Println("(proxy): Job session proxySession exists already. Reopen it.")
		if js, err = sm.OpenJobSession(jsName); err != nil {
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
			fmt.Println("Error during GetAllJobs(): %s", err)
		} else {
			log.Println("No job in that state found!")
		}
	} else {
		log.Printf("amount of matching jobs %d\n", len(job))
		if len(job) >= 1 {
			ji := make([]drmaa2.JobInfo, 0, 500)
			for i, _ := range job {
				jinfo, _ := job[i].GetJobInfo()
				ji = append(ji, *jinfo)
			}
			return ji
		}
	}
	return nil
}
