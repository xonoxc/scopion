package dockerclient

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerService struct {
	cli *client.Client
}

func New() (*DockerService, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	return &DockerService{cli: cli}, nil
}

func (d *DockerService) RunContainer(ctx context.Context, spec ContainerSpec) (*ContainerResult, error) {
	reader, err := d.cli.ImagePull(ctx, spec.Image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}

	io.Copy(io.Discard, reader)
	reader.Close()

	exposed := nat.PortSet{}
	bindings := nat.PortMap{}

	for cPort, hPort := range spec.Ports {
		port := nat.Port(cPort)
		exposed[port] = struct{}{}
		bindings[port] = []nat.PortBinding{
			{HostIP: "127.0.0.1", HostPort: hPort},
		}
	}

	resp, err := d.cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:        spec.Image,
			Env:          spec.Env,
			ExposedPorts: exposed,
		},
		&container.HostConfig{
			PortBindings: bindings,
		},
		nil,
		nil,
		spec.Name,
	)
	if err != nil {
		return nil, err
	}

	if err := d.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	return &ContainerResult{
		ID:    resp.ID,
		Name:  spec.Name,
		Ports: spec.Ports,
	}, nil
}
