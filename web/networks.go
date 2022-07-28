package web

import "context"

type (
	NetworkID string

	Network struct {
		ID   NetworkID `json:"id"`
		Name string    `json:"name"`
	}

	NetworkProvider interface {
		GetNetwork(NetworkID, context.Context) (*Network, error)
	}
)
