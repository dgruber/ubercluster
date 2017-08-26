package fake

import (
	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"io"
	"time"
)

// FakeDocker implements DockerInterface
type FakeDocker struct {
	containerList map[string]*container.Config
}

func NewFakeDocker() *FakeDocker {
	return &FakeDocker{
		containerList: make(map[string]*container.Config),
	}
}

func (f *FakeDocker) ContainerCreate(config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	f.containerList[containerName] = config
	return container.ContainerCreateCreatedBody{}, nil
}

func (f *FakeDocker) ContainerStart(containerID string, options dtypes.ContainerStartOptions) error {
	return nil
}

func (f *FakeDocker) ContainerPause(containerID string) error {
	return nil
}

func (f *FakeDocker) ContainerUnpause(containerID string) error {
	return nil
}

func (f *FakeDocker) ContainerStop(containerID string, timeout *time.Duration) error {
	return nil
}

func (f *FakeDocker) ContainerList(clo dtypes.ContainerListOptions) ([]dtypes.Container, error) {
	return nil, nil
}

func (f *FakeDocker) ContainerInspect(containerID string) (dtypes.ContainerJSON, error) {
	return dtypes.ContainerJSON{}, nil
}

func (f *FakeDocker) ImageList(dtypes.ImageListOptions) ([]dtypes.ImageSummary, error) {
	return []dtypes.ImageSummary{
		dtypes.ImageSummary{RepoDigests: []string{"golang/latest"}},
		dtypes.ImageSummary{RepoDigests: []string{"google/golang"}},
	}, nil
}

func (f *FakeDocker) ImagePull(refStr string, options dtypes.ImagePullOptions) (io.ReadCloser, error) {
	return nil, nil
}

func (f *FakeDocker) ClientVersion() string {
	return "1.0.0"
}
