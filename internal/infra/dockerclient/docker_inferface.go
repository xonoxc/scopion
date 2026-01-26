package dockerclient

import "context"

type ContainerSpec struct {
	Image string
	Name  string
	Env   []string
	Ports map[string]string
}

type ContainerResult struct {
	ID    string
	Name  string
	Ports map[string]string
}

type Service interface {
	RunContainer(ctx context.Context, spec ContainerSpec) (*ContainerResult, error)
}
