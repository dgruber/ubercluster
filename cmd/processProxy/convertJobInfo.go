package main

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/ubercluster/pkg/types"
)

func ConvertJobInfo(d drmaa2interface.JobInfo) *types.JobInfo {
	var t types.JobInfo
	t.Id = d.ID
	t.ExitStatus = d.ExitStatus
	t.TerminatingSignal = d.TerminatingSignal
	t.Annotation = d.Annotation
	t.State = (types.JobState)(d.State)
	t.SubState = d.SubState
	t.AllocatedMachines = make([]string, len(d.AllocatedMachines))
	copy(t.AllocatedMachines, d.AllocatedMachines)
	t.SubmissionMachine = d.SubmissionMachine
	t.JobOwner = d.JobOwner
	t.Slots = d.Slots
	t.QueueName = d.QueueName
	t.WallclockTime = d.WallclockTime
	t.CPUTime = d.CPUTime
	t.SubmissionTime = d.SubmissionTime
	t.DispatchTime = d.DispatchTime
	t.FinishTime = d.FinishTime
	return &t
}
