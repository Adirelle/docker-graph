package api

import "context"

type (
	ContainerID string

	Container struct {
		ID       ContainerID `json:"id"`
		Name     string      `json:"name"`
		Status   string      `json:"status"`
		Healty   string      `json:"healty"`
		Project  string      `json:"project,omitempty"`
		Service  string      `json:"service,omitempty"`
		Networks []NetworkID `json:"networks,omitempty"`
		Mounts   []Mount     `json:"mounts,omitempty"`
	}

	Mount struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
		ReadWrite   bool   `json:"read_write"`
	}

	ContainerProvider interface {
		ListContainerIDs(context.Context) ([]ContainerID, error)
		GetContainer(ContainerID, context.Context) (*Container, error)
	}
)
