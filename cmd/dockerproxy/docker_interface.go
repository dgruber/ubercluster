package main

import (
	"fmt"
	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"io"
	"time"
)

type DockerInterface interface {
	ContainerCreate(config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(containerID string, options dtypes.ContainerStartOptions) error
	ContainerPause(containerID string) error
	ContainerUnpause(containerID string) error
	ContainerStop(containerID string, timeout *time.Duration) error
	ContainerList(options dtypes.ContainerListOptions) ([]dtypes.Container, error)
	ContainerInspect(containerID string) (dtypes.ContainerJSON, error)
	ImageList(dtypes.ImageListOptions) ([]dtypes.ImageSummary, error)
	ImagePull(refStr string, options dtypes.ImagePullOptions) (io.ReadCloser, error)
	ClientVersion() string
}

type Docker struct {
	cli *client.Client
	ctx context.Context
}

func (d *Docker) ContainerCreate(config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	return d.cli.ContainerCreate(d.ctx, config, hostConfig, networkingConfig, containerName)
}

func (d *Docker) ContainerStart(containerID string, options dtypes.ContainerStartOptions) error {
	return d.cli.ContainerStart(d.ctx, containerID, options)
}

func (d *Docker) ContainerPause(containerID string) error {
	return d.cli.ContainerPause(d.ctx, containerID)
}

func (d *Docker) ContainerUnpause(containerID string) error {
	return d.cli.ContainerPause(d.ctx, containerID)
}

func (d *Docker) ContainerStop(containerID string, timeout *time.Duration) error {
	return d.cli.ContainerStop(d.ctx, containerID, timeout)
}

func (d *Docker) ContainerList(options dtypes.ContainerListOptions) ([]dtypes.Container, error) {
	return d.cli.ContainerList(d.ctx, options)
}

func (d *Docker) ContainerInspect(containerID string) (dtypes.ContainerJSON, error) {
	return d.cli.ContainerInspect(d.ctx, containerID)
}

func (d *Docker) ImagePull(refStr string, options dtypes.ImagePullOptions) (io.ReadCloser, error) {
	return d.cli.ImagePull(d.ctx, refStr, options)
}

func (d *Docker) ImageList(ilo dtypes.ImageListOptions) ([]dtypes.ImageSummary, error) {
	return d.cli.ImageList(d.ctx, ilo)
}

func (d *Docker) ClientVersion() string {
	return d.cli.ClientVersion()
}

func NewDocker(ctx context.Context) (DockerInterface, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("Couldn't create Docker proxy: %s", err.Error())
	}
	return &Docker{cli: cli, ctx: ctx}, nil
}
