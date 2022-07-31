package connections

import (
	"io"

	"github.com/docker/docker/client"
)

type (
	Connection interface {
		client.ContainerAPIClient
		client.NetworkAPIClient
		client.VolumeAPIClient
		client.SystemAPIClient
		io.Closer
	}

	Factory interface {
		CreateConn() (Connection, error)
	}
)
