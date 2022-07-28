package web

import "context"

type (
	ContainerID string

	Container struct {
		ID       ContainerID
		Name     string
		Status   string
		Networks []NetworkID
		Mounts   []Mount
	}

	Mount struct {
		Name        string
		Type        string
		Source      string
		Destination string
		ReadWrite   bool
	}

	ContainerProvider interface {
		ListContainerIDs(context.Context) ([]ContainerID, error)
		GetContainer(ContainerID, context.Context) (*Container, error)
	}
)
