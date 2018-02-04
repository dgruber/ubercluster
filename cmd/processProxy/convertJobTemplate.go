package main

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/ubercluster/pkg/types"
)

func copyMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func ConvertJobTemplate(u types.JobTemplate) (jt drmaa2interface.JobTemplate) {
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
	jt.ReservationID = u.ReservationId
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
	jt.AccountingID = u.AccountingId
	return jt
}
