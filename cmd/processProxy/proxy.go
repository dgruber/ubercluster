package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"
	"github.com/dgruber/ubercluster/pkg/types"
)

type Proxy struct {
	SessionManager *drmaa2os.SessionManager
	JobSession     drmaa2interface.JobSession
}

func NewProxy() Proxy {
	sm, err := drmaa2os.NewDefaultSessionManager("ucProxy.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create SessionManager for processes (%s).\n", err.Error())
		os.Exit(1)
	}
	js, errCreate := sm.CreateJobSession(SESSION_NAME, "")
	if errCreate != nil {
		var errOpen error
		js, errOpen = sm.OpenJobSession(SESSION_NAME)
		if errOpen != nil {
			fmt.Fprintf(os.Stderr, "Could not create or open JobSession for processes (%s).\n", errCreate.Error())
			os.Exit(1)
		}
	}
	return Proxy{
		SessionManager: sm,
		JobSession:     js,
	}
}

// RunJob creates a process.
func (p *Proxy) RunJob(template types.JobTemplate) (string, error) {
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

	job, err := p.JobSession.RunJob(ConvertJobTemplate(template))
	if err != nil {
		return "", err
	}

	return job.GetID(), nil
}

func jobByID(p *Proxy, jobid string) (drmaa2interface.Job, error) {
	filter := drmaa2interface.CreateJobInfo()
	filter.ID = jobid
	jobs, err := p.JobSession.GetJobs(filter)
	if err != nil {
		return nil, err
	}
	if len(jobs) < 1 {
		return nil, errors.New("couldn't fetch job (job list length 0)")
	}
	return jobs[0], nil
}

// JobOperation changes the state of a job in the system.
func (p *Proxy) JobOperation(jobsessionname, operation, jobid string) (out string, err error) {
	job, err := jobByID(p, jobid)
	if err != nil {
		return "", err
	}

	switch operation {
	case "suspend":
		if opErr := job.Suspend(); opErr != nil {
			err = opErr
		} else {
			out = "Suspended Job"
		}
	case "resume":
		if opErr := job.Resume(); opErr != nil {
			err = opErr
		} else {
			out = "Resumed Job"
		}
	case "terminate":
		if opErr := job.Terminate(); opErr != nil {
			err = opErr
		} else {
			out = "Terminated Job"
		}
		// hold and resume not supported for processes
	default:
		log.Println("JobOperation unknown operation ", operation)
		err = errors.New("Unknown operation: " + operation)
	}
	return out, err
}

// GetJobInfosByFilter
func (p *Proxy) GetJobInfosByFilter(filtered bool, filter types.JobInfo) []types.JobInfo {
	if filtered == false {
		jobs, err := p.JobSession.GetJobs(drmaa2interface.CreateJobInfo())
		if err != nil {
			fmt.Printf("GetJobInfosByFilter(): %s\n", err.Error())
			return nil
		}
		jobInfos := make([]types.JobInfo, len(jobs))
		for _, job := range jobs {
			j := p.GetJobInfo(job.GetID())
			if j != nil {
				jobInfos = append(jobInfos, *j)
			}
		}
		return jobInfos
	}
	fmt.Println("GetJobInfosByFilter() with filter not implemented")
	return nil
}

// GetJobInfo returns information about a job.
func (p *Proxy) GetJobInfo(jobid string) *types.JobInfo {
	job, err := jobByID(p, jobid)
	if err != nil {
		fmt.Printf("GetJobInfo(): %s\n", err.Error())
		return nil
	}
	jobInfo, errJI := job.GetJobInfo()
	if errJI != nil {
		return nil
	}
	return ConvertJobInfo(jobInfo)
}

// GetAllMachines
func (p *Proxy) GetAllMachines(machines []string) ([]types.Machine, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("can not get hostname of machine: %s", err)
	}
	if machines == nil {
		// TODO Machine details
		return []types.Machine{types.Machine{Name: hostname}}, nil
	}
	for i := range machines {
		if machines[i] == hostname {
			// TODO Machine details
			return []types.Machine{types.Machine{Name: hostname}}, nil
		}
	}
	return []types.Machine{}, nil
}

// GetAllQueues
func (p *Proxy) GetAllQueues(queues []string) ([]types.Queue, error) {
	q := types.Queue{
		Name: "os",
	}
	if queues == nil {
		return []types.Queue{q}, nil
	}
	for i := range queues {
		if queues[i] == "os" {
			return []types.Queue{q}, nil
		}
	}
	return []types.Queue{}, nil
}

// GetAllSessions just returns the by the proxy created job session.
func (p *Proxy) GetAllSessions(session []string) ([]string, error) {
	if session == nil {
		return []string{SESSION_NAME}, nil
	}
	for i := range session {
		if session[i] == SESSION_NAME {
			return []string{SESSION_NAME}, nil
		}
	}
	return []string{}, nil
}

// GetAllCategories returns nothing since there are no job categories.
func (p *Proxy) GetAllCategories() ([]string, error) {
	return []string{}, nil
}

// DRMSVersion returns the version of the DRMAA implementation.
func (p *Proxy) DRMSVersion() string {
	version, err := p.SessionManager.GetDrmsVersion()
	if err != nil {
		return "unknown"
	}
	return version.String()
}

// DRMSName returns the process manager implementation.
func (p *Proxy) DRMSName() string {
	name, err := p.SessionManager.GetDrmsName()
	if err != nil {
		return "unknown"
	}
	return name
}

// DRMSLoad returns the load of the host.
func (p *Proxy) DRMSLoad() float64 {
	return 0.5
}
