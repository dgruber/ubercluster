package main

import (
	"github.com/dgruber/go-cfclient"
	"github.com/dgruber/ubercluster/pkg/types"
)

func TransformTaskInJobInfo(task cfclient.Task) types.JobInfo {
	info := types.JobInfo{
		Id:             task.GUID,
		SubmissionTime: task.CreatedAt,
		DispatchTime:   task.CreatedAt,
	}
	switch task.State {
	case "SUCCEEDED":
		info.State = types.Done
		info.FinishTime = task.UpdatedAt
	case "FAILED":
		info.State = types.Failed
		info.FinishTime = task.UpdatedAt
	case "RUNNING":
		info.State = types.Running
	default:
		info.State = types.Undetermined
	}
	return info
}

func TransformTasksInJobInfo(task []cfclient.Task) []types.JobInfo {
	ji := make([]types.JobInfo, 0, len(task))
	for i := range task {
		ji = append(ji, TransformTaskInJobInfo(task[i]))
	}
	return ji
}

func (cp *CFProxy) getJobInfos() []types.JobInfo {
	tasks, err := cp.client.ListTasks()
	if err != nil {
		return nil
	}
	return TransformTasksInJobInfo(tasks)
}

func (cp *CFProxy) getJobInfo(guid string) *types.JobInfo {
	task, err := cp.client.TaskByGuid(guid)
	if err != nil {
		return nil
	}
	ji := TransformTaskInJobInfo(task)
	return &ji
}
