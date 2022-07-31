package connections

import (
	"log"

	"github.com/docker/docker/client"
)

type (
	BasicFactory []client.Opt
)

var (
	_ Factory    = (BasicFactory)(nil)
	_ Connection = (*client.Client)(nil)
)

func MakeBasicFactory(opts ...client.Opt) BasicFactory {
	return BasicFactory(opts)
}

func (f BasicFactory) CreateConn() (Connection, error) {
	client, err := client.NewClientWithOpts(f...)
	if err != nil {
		return nil, err
	}
	log.Println("Opened connection")
	return client, nil
}
