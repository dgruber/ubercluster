package main

import (
	"errors"
	"github.com/dgruber/go-cfclient"
	"github.com/dgruber/ubercluster/pkg/types"
	"log"
)

// CFProxy implements the proxy interface
type CFProxy struct {
	client *cfclient.Client
}

func New() (*CFProxy, error) {
	config, errConfig := discoverConfig()
	if errConfig != nil {
		return nil, errConfig
	}
	client, errClient := cfclient.NewClient(config)
	if errClient != nil {
		return nil, errClient
	}
	return &CFProxy{
		client: client,
	}, nil
}

func (cp *CFProxy) RunJob(template types.JobTemplate) (jobid string, err error) {
	if template.JobCategory == "" {
		return "", errors.New("No jobcategory (app name) requested!")
	}
	return cp.runTask(template)
}

func (cp *CFProxy) JobOperation(jobsessionname, operation, jobid string) (out string, err error) {
	switch operation {
	case "suspend":
		err = errors.New("Unsupported operation: \"suspend\"")
	case "resume":
		err = errors.New("Unsupported operation: \"resume\"")
	case "terminate":
		err = cp.client.TerminateTask(jobid)
		if err != nil {
			return "Failed", err
		}
		return "Terminated job", nil
	default:
		log.Printf("JobOperation unknown operation: %s", operation)
		err = errors.New("Unknown operation: " + operation)
	}
	return out, err
}

func (cp *CFProxy) GetJobInfosByFilter(filtered bool, filter types.JobInfo) []types.JobInfo {
	return cp.getJobInfos()
}

func (cp *CFProxy) GetJobInfo(jobid string) *types.JobInfo {
	return cp.getJobInfo(jobid)
}

func (cp *CFProxy) GetAllMachines(machines []string) ([]types.Machine, error) {
	return nil, nil
}

func (cp *CFProxy) GetAllQueues(queues []string) ([]types.Queue, error) {
	return nil, nil
}

func (cp *CFProxy) GetAllSessions(session []string) ([]string, error) {
	return nil, nil
}

func (cp *CFProxy) GetAllCategories() ([]string, error) {
	return cp.listApps()
}

func (cp *CFProxy) DRMSVersion() string {
	return "0.1"
}

func (cp *CFProxy) DRMSName() string {
	return "Cloud Foundry Tasks"
}

func (cp *CFProxy) DRMSLoad() float64 {
	return 0.5
}
