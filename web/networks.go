package web

import "context"

type (
	NetworkID string

	Network struct {
		ID   NetworkID
		Name string
	}

	NetworkProvider interface {
		GetNetwork(NetworkID, context.Context) (*Network, error)
	}
)
