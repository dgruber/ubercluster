package main

import (
	"errors"
	"github.com/dgruber/go-cfclient"
	"os"
)

func discoverConfig() (*cfclient.Config, error) {
	address, existsCF := os.LookupEnv("CF_TARGET")
	if existsCF == false {
		return nil, errors.New("CF_TARGET env not set")
	}
	name, existsName := os.LookupEnv("NAME")
	if existsName == false {
		return nil, errors.New("NAME env not set")
	}
	password, existsPassword := os.LookupEnv("PASSWORD")
	if existsPassword == false {
		return nil, errors.New("PASSWORD env not set")
	}
	c := &cfclient.Config{
		ApiAddress:        address,
		Username:          name,
		Password:          password,
		SkipSslValidation: true,
	}
	return c, nil
}
