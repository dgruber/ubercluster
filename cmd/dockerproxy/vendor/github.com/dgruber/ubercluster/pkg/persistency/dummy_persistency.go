package persistency

import (
	"github.com/dgruber/ubercluster/pkg/types"
	"log"
)

// DummyPersistenxy implements the PersistencyImplementer interface
// but does not do anyhthing when one of the methods are called.
type DummyPersistency struct {
}

func (dp *DummyPersistency) SaveJobTemplate(jobid string, jt types.JobTemplate) error {
	log.Println("SaveJobTemplate called")
	return nil
}

func (dp *DummyPersistency) SaveJobInfo(jobid string, ji types.JobInfo) error {
	log.Println("SaveJobInfo called")
	return nil
}
