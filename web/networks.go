package web

import "context"

type (
	NetworkID string

	Network struct {
		ID   NetworkID
		Name string
	}

	NetworkProvider interface {
		ListNetworkIDs(context.Context) ([]NetworkID, error)
		GetNetwork(NetworkID, context.Context) (*Network, error)
	}
)
