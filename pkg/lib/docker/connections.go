package docker

import (
	"io"
	"log"

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

	ConnectionFactory interface {
		CreateConn() (Connection, error)
	}

	BasicConnectionFactory []client.Opt
)

var (
	_ ConnectionFactory = (BasicConnectionFactory)(nil)
	_ Connection        = (*client.Client)(nil)
)

func MakeBasicConnectionFactory(opts ...client.Opt) BasicConnectionFactory {
	return BasicConnectionFactory(opts)
}

func (f BasicConnectionFactory) CreateConn() (Connection, error) {
	client, err := client.NewClientWithOpts(f...)
	if err != nil {
		return nil, err
	}
	log.Println("Opened connection")
	return client, nil
}
