package main

import (
	"fmt"
)

func (p *Proxy) stopContainer(id string) error {
	err := p.client.ContainerStop(id, nil)
	p.client.ContainerPause(id)
	if err != nil {
		return fmt.Errorf("Can not stop container: %s", err.Error())
	}
	return nil
}

func (p *Proxy) pauseContainer(id string) error {
	err := p.client.ContainerPause(id)
	if err != nil {
		return fmt.Errorf("Can not pause container: %s", err.Error())
	}
	return nil
}

func (p *Proxy) unpauseContainer(id string) error {
	err := p.client.ContainerUnpause(id)
	if err != nil {
		return fmt.Errorf("Can not unpause container: %s", err.Error())
	}
	return nil
}
