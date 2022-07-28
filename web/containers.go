package web

import "context"

type (
	ContainerID string

	Container struct {
		ID       ContainerID
		Name     string
		Status   string
		Networks []NetworkID
	}

	ContainerProvider interface {
		ListContainerIDs(context.Context) ([]ContainerID, error)
		GetContainer(ContainerID, context.Context) (*Container, error)
	}
)
