package main

import (
	"fmt"
	dtypes "github.com/docker/docker/api/types"
)

func (p *Proxy) listDockerImages() ([]string, error) {
	images, err := p.client.ImageList(dtypes.ImageListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Error during listing Docker images: %s", err)
	}
	dockerImages := make([]string, 0, len(images))
	for _, image := range images {
		dockerImages = append(dockerImages, image.RepoTags...)
	}
	return dockerImages, nil
}
