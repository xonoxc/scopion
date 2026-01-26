package dockerclient

import (
	"context"

	docker "github.com/fsouza/go-dockerclient"
)

type DockerService struct {
	cli *docker.Client
}

func New() (*DockerService, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	return &DockerService{
		cli: client,
	}, nil
}

func (d *DockerService) RunContainer(ctx context.Context, spec ContainerSpec) (*ContainerResult, error) {
	err := d.cli.PullImage(
		docker.PullImageOptions{
			Repository: spec.Image,
			Context:    ctx,
		},
		docker.AuthConfiguration{},
	)
	if err != nil {
		return nil, err
	}

	exposedPorts := map[docker.Port]struct{}{}
	portBindings := map[docker.Port][]docker.PortBinding{}

	for cPort, hPort := range spec.Ports {
		port := docker.Port(cPort)
		exposedPorts[port] = struct{}{}
		portBindings[port] = []docker.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: hPort,
			},
		}
	}

	container, err := d.cli.CreateContainer(docker.CreateContainerOptions{
		Name:    spec.Name,
		Context: ctx,
		Config: &docker.Config{
			Image:        spec.Image,
			Env:          spec.Env,
			ExposedPorts: exposedPorts,
		},
		HostConfig: &docker.HostConfig{
			PortBindings: portBindings,
		},
	})
	if err != nil {
		return nil, err
	}

	if err := d.cli.StartContainer(container.ID, nil); err != nil {
		return nil, err
	}

	return &ContainerResult{
		ID:    container.ID,
		Name:  spec.Name,
		Ports: spec.Ports,
	}, nil
}
