package main

import (
	"errors"
	"fmt"
	"github.com/dgruber/go-cfclient"
	"github.com/dgruber/ubercluster/pkg/types"
	"strings"
)

func (cp *CFProxy) findAppGUID(appName string) (string, error) {
	apps, err := cp.client.ListApps()
	if err != nil {
		return "", fmt.Errorf("Could not find app: %s", err.Error())
	}
	for i := range apps {
		if apps[i].Name == appName {
			return apps[i].Guid, nil
		}
	}
	return "", fmt.Errorf("Could not find app %s", appName)
}

func createTaskRequest(jt types.JobTemplate, guid string) (tr cfclient.TaskRequest, err error) {
	tr.Command = jt.RemoteCommand
	if len(jt.Args) > 0 {
		tr.Command += " " + strings.Join(jt.Args, " ")
	}
	tr.Name = jt.JobName
	if jt.MinPhysMemory > 0 {
		tr.MemoryInMegabyte = int(jt.MinPhysMemory / 1024)
	}
	tr.DiskInMegabyte = 0
	tr.DropletGUID = guid

	return tr, nil
}

func (cp *CFProxy) runTask(jt types.JobTemplate) (string, error) {
	if jt.JobCategory == "" {
		return "", errors.New("No job category (app) requested.")
	}
	guid, errGUID := cp.findAppGUID(jt.JobCategory)
	if errGUID != nil {
		return "", errGUID
	}

	treq, errTR := createTaskRequest(jt, guid)
	if errTR != nil {
		return "", errTR
	}
	task, errCreate := cp.client.CreateTask(treq)
	if errCreate != nil {
		return "", errCreate
	}
	return task.GUID, nil
}
