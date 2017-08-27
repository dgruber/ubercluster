package main

import (
	"errors"
	"github.com/dgruber/ubercluster/pkg/types"
	"golang.org/x/net/context"
	"log"
)

// Proxy implements the proxy interface
type Proxy struct {
	client DockerInterface
	ctx    context.Context
	config *DockerConfig
}

func New() (*Proxy, error) {
	config, errConfig := discoverConfig()
	if errConfig != nil {
		return nil, errConfig
	}
	dockerClient, err := NewDocker(context.Background())
	if err != nil {
		return nil, err
	}
	return NewProxy(dockerClient, config), nil
}

func NewProxy(client DockerInterface, config *DockerConfig) *Proxy {
	return &Proxy{
		client: client,
		ctx:    context.Background(),
		config: config,
	}
}

func (p *Proxy) RunJob(template types.JobTemplate) (jobid string, err error) {
	if template.JobCategory == "" {
		return "", errors.New("No jobcategory (docker image name) requested!")
	}
	return p.runTask(template)
}

func (p *Proxy) JobOperation(jobsessionname, operation, jobid string) (out string, err error) {
	switch operation {
	case "suspend":
		err := p.pauseContainer(jobid)
		if err != nil {
			return "", err
		}
		return "Suspended job", nil
	case "resume":
		err := p.unpauseContainer(jobid)
		if err != nil {
			return "", err
		}
		return "Resumed job", nil
	case "terminate":
		err := p.stopContainer(jobid)
		if err != nil {
			return out, err
		}
		return "Terminated job", nil
	default:
		log.Printf("JobOperation unknown operation: %s", operation)
		err = errors.New("Unknown operation: " + operation)
	}
	return out, err
}

func (p *Proxy) GetJobInfosByFilter(filtered bool, filter types.JobInfo) []types.JobInfo {
	return p.getJobInfos(filtered, filter)
}

func (p *Proxy) GetJobInfo(jobid string) *types.JobInfo {
	return p.getJobInfo(jobid)
}

func (p *Proxy) GetAllMachines(machines []string) ([]types.Machine, error) {
	return []types.Machine{LocalhostToMachine()}, nil
}

func (p *Proxy) GetAllQueues(queues []string) ([]types.Queue, error) {
	return []types.Queue{}, nil
}

func (p *Proxy) GetAllSessions(session []string) ([]string, error) {
	return []string{}, nil
}

func (p *Proxy) GetAllCategories() ([]string, error) {
	return p.listDockerImages()
}

func (p *Proxy) DRMSVersion() string {
	return p.client.ClientVersion()
}

func (p *Proxy) DRMSName() string {
	return "Docker"
}

func (p *Proxy) DRMSLoad() float64 {
	return 0.5
}
