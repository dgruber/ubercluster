package main

import (
	"errors"
	"fmt"
	"github.com/dgruber/ubercluster/pkg/types"
	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

// https://github.com/moby/moby/blob/master/api/types/container/config.go
func jobTemplateToContainerConfig(jt types.JobTemplate) (*container.Config, *container.HostConfig, error) {
	var cc container.Config
	var hc container.HostConfig

	cc.Image = jt.JobCategory
	cc.WorkingDir = jt.WorkingDirectory

	if len(jt.CandidateMachines) == 1 {
		cc.Hostname = jt.CandidateMachines[0]
	}

	if jt.RemoteCommand != "#nocommand#" {
		cmdSlice := []string{jt.RemoteCommand}
		cmdSlice = append(cmdSlice, jt.Args...)
		cc.Cmd = cmdSlice
	}

	hc.PublishAllPorts = true

	return &cc, &hc, nil
}

func pullImage(client DockerInterface, image string) error {
	_, err := client.ImagePull(image, dtypes.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("Error during pulling image: %s", err.Error())
	}
	return nil
}

func imageExists(client DockerInterface, image string) (bool, error) {
	summary, err := client.ImageList(dtypes.ImageListOptions{})
	if err != nil {
		return false, fmt.Errorf("Error during listing images: %s", summary)
	}
	for _, i := range summary {
		for _, tag := range i.RepoTags {
			if tag == image {
				return true, nil
			}
		}
	}
	return false, nil
}

func (p *Proxy) runTask(jt types.JobTemplate) (string, error) {
	if jt.JobCategory == "" {
		return "", errors.New("No job category (app) requested.")
	}

	cc, hc, err := jobTemplateToContainerConfig(jt)
	if err != nil {
		return "", fmt.Errorf("Can not run task (%s).", err.Error())
	}

	if p.config.AllowImagePull {
		exists, _ := imageExists(p.client, jt.JobCategory)
		if exists == false {
			fmt.Printf("Pulling image %s\n", jt.JobCategory)
			errPull := pullImage(p.client, jt.JobCategory)
			if errPull != nil {
				return "", errPull
			}
		}
	}

	resp, err := p.client.ContainerCreate(cc, hc, nil, "")
	if err != nil {
		return "", fmt.Errorf("Error during container creation: %s", err.Error())
	}

	err = p.client.ContainerStart(resp.ID, dtypes.ContainerStartOptions{})
	if err != nil {
		return "", fmt.Errorf("Error during staring container %s: %s", resp.ID, err.Error())
	}
	return resp.ID, nil
}
