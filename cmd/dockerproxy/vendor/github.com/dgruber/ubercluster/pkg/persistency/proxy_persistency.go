package persistency

import (
	"github.com/dgruber/ubercluster/pkg/types"
)

// PersistencyImplementer is an interface which contains
// all functions required for making elements (job templates /
// states etc.) persistent on some endpoints (databases,
// filesystem / cloud storage etc.). This functionality is
// required when the DRM does not have support for it or when
// more advanced operations (moving submitted jobs between
// clusters) are going to be implemented.
type PersistencyImplementer interface {
	// SaveJobTemplate makes the JobTemplate persistent. This is done after
	// a job was submitted by the client.
	SaveJobTemplate(jobid string, jinfo types.JobTemplate) error
	// SaveJobInfo makes the JobInfo object persistent. This is usually done
	// in intervalls or when the user requests a JobInfo object or when the
	// job is reaped from the DRM:
	SaveJobInfo(jobid string, jinfo types.JobInfo) error
}
