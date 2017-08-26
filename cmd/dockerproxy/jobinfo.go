package main

import (
	"fmt"
	"github.com/dgruber/ubercluster/pkg/types"
	dtypes "github.com/docker/docker/api/types"
	_ "github.com/docker/docker/api/types/container"
	"golang.org/x/net/context"
	"os"
	"os/user"
	"time"
)

func (p *Proxy) getJobInfo(jobid string) *types.JobInfo {
	containers, err := getAllContainers(p.client, p.ctx)
	if err != nil {
		return nil
	}
	ctr := findContainer(containers, jobid)
	if ctr == nil {
		return nil
	}
	return convertContainer(p.client, p.ctx, ctr)
}

func (p *Proxy) getJobInfos(filtered bool, filter types.JobInfo) []types.JobInfo {
	containers, err := getAllContainers(p.client, p.ctx)
	if err != nil {
		return nil
	}
	return convertAllContainers(p.client, p.ctx, containers)
}

func getAllContainers(client DockerInterface, ctx context.Context) ([]dtypes.Container, error) {
	containers, err := client.ContainerList(dtypes.ContainerListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Error listing containers: %s", err.Error())
	}
	return containers, nil
}

func findContainer(containers []dtypes.Container, jobid string) *dtypes.Container {
	for _, ctr := range containers {
		if ctr.ID == jobid {
			return &ctr
		}
	}
	return nil
}

func convertAllContainers(client DockerInterface, ctx context.Context, containers []dtypes.Container) []types.JobInfo {
	jis := make([]types.JobInfo, 0, len(containers))
	for i := range containers {
		jis = append(jis, *convertContainer(client, ctx, &containers[i]))
	}
	return jis
}

func convertContainer(client DockerInterface, ctx context.Context, ctr *dtypes.Container) *types.JobInfo {
	jobowner := ""
	if user, err := user.Current(); err == nil {
		jobowner = user.Username
	}
	hostname, _ := os.Hostname()

	var exitStatus int
	var terminationSignal string
	var status types.JobState
	status = types.Undetermined
	var finishTime time.Time

	if ctrInspect, err := client.ContainerInspect(ctr.ID); err != nil {
		hostname = ctrInspect.HostnamePath
		exitStatus = ctrInspect.State.ExitCode
		if ctrInspect.State.Paused {
			status = types.Suspended
		}
		if ctrInspect.State.Restarting || ctrInspect.State.Running {
			status = types.Running
		}
		finishTime, _ = time.Parse("2015-01-06T15:47:32.080254511Z", ctrInspect.State.FinishedAt)
	}

	ji := types.JobInfo{
		Id:                ctr.ID,
		DispatchTime:      time.Unix(ctr.Created, 0),
		JobOwner:          jobowner,
		Slots:             1,
		AllocatedMachines: []string{hostname},
		ExitStatus:        exitStatus,
		TerminatingSignal: terminationSignal,
		State:             status,
		FinishTime:        finishTime,
	}
	return &ji
}
