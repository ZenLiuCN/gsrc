package dsock

import (
	. "github.com/docker/docker/client"
)

type Docker struct {
	APIClient
}

func NewDockerClient(ver string) (dc *Docker, err error) {
	dc = new(Docker)
	dc.APIClient, err = NewClientWithOpts(WithVersion(ver), FromEnv)
	return
}
