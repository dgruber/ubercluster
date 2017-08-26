package main

import (
	"errors"
	"os"
	"strconv"
)

type DockerConfig struct {
	AllowImagePull bool
}

func discoverConfig() (*DockerConfig, error) {
	pull, existsImagePull := os.LookupEnv("UC_DOCKER_IMAGE_PULL")
	if existsImagePull == false {
		pull = "true"
	}
	allowPull, errParseBool := strconv.ParseBool(pull)
	if errParseBool != nil {
		return nil, errors.New("UC_DOCKER_IMAGE_PULL can't be parsed (must be set to true or false)")
	}

	c := &DockerConfig{
		AllowImagePull: allowPull,
	}
	return c, nil
}
