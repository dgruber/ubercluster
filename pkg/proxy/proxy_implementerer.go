package proxy

import (
	"github.com/dgruber/ubercluster/pkg/types"
)

// ProxyImplementer interface specified functions required to interface
// a ubercluster proxy. Those functions are called in the standard
// http request handlers.
type ProxyImplementer interface {
	GetJobInfosByFilter(filtered bool, filter types.JobInfo) []types.JobInfo
	GetJobInfo(jobid string) *types.JobInfo
	GetAllMachines(machines []string) ([]types.Machine, error)
	GetAllQueues(queues []string) ([]types.Queue, error)
	GetAllCategories() ([]string, error)
	GetAllSessions(session []string) ([]string, error)
	DRMSVersion() string
	DRMSName() string
	RunJob(template types.JobTemplate) (string, error)
	JobOperation(jobsessionname, operation, jobid string) (string, error)
	DRMSLoad() float64
}
