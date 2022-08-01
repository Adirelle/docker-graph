package connections

import (
	"context"

	log "github.com/inconshreveable/log15"

	"github.com/docker/docker/client"
)

type (
	BasicFactory []client.Opt
)

var (
	_ Factory    = (BasicFactory)(nil)
	_ Connection = (*client.Client)(nil)

	Log = log.New()
)

func MakeBasicFactory(opts ...client.Opt) BasicFactory {
	return BasicFactory(opts)
}

func (f BasicFactory) CreateConn() (Connection, error) {
	client, err := client.NewClientWithOpts(f...)
	if err != nil {
		return nil, err
	}
	ping, err := client.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	log.Info("opened connection", log.Ctx{
		"host":            client.DaemonHost(),
		"api_version":     ping.APIVersion,
		"builder_version": ping.BuilderVersion,
		"os_type":         ping.OSType,
	})
	return client, nil
}
