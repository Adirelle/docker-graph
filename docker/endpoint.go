package docker

import (
	"context"

	"github.com/adirelle/docker-graph/web"
	"github.com/docker/docker/api/types"
)

type (
	Endpoint struct {
		connFactory ConnectionFactory
	}
)

var (
	_ web.ContainerProvider = (*Endpoint)(nil)
	_ web.NetworkProvider   = (*Endpoint)(nil)
)

func NewEndpoint(connFactory ConnectionFactory) *Endpoint {
	return &Endpoint{connFactory}
}

func (e *Endpoint) ListContainerIDs(ctx context.Context) (ids []web.ContainerID, err error) {
	var conn Connection
	if conn, err = e.connFactory.CreateConn(); err != nil {
		return
	}
	defer conn.Close()
	var containers []types.Container
	if containers, err = conn.ContainerList(ctx, types.ContainerListOptions{}); err == nil {
		for _, container := range containers {
			ids = append(ids, web.ContainerID(container.ID))
		}
	}
	return
}

func (e *Endpoint) GetContainer(id web.ContainerID, ctx context.Context) (ctn *web.Container, err error) {
	var conn Connection
	if conn, err = e.connFactory.CreateConn(); err != nil {
		return
	}
	defer conn.Close()
	var data types.ContainerJSON
	if data, err = conn.ContainerInspect(ctx, string(id)); err == nil {
		ctn = &web.Container{ID: id, Name: data.Name, Status: data.State.Status}
		for _, net := range data.NetworkSettings.Networks {
			ctn.Networks = append(ctn.Networks, web.NetworkID(net.NetworkID))
		}
	}
	return
}

func (e *Endpoint) ListNetworkIDs(ctx context.Context) (ids []web.NetworkID, err error) {
	var conn Connection
	if conn, err = e.connFactory.CreateConn(); err != nil {
		return
	}
	defer conn.Close()
	var networks []types.NetworkResource
	if networks, err = conn.NetworkList(ctx, types.NetworkListOptions{}); err == nil {
		for _, network := range networks {
			ids = append(ids, web.NetworkID(network.ID))
		}
	}
	return
}

func (e *Endpoint) GetNetwork(id web.NetworkID, ctx context.Context) (net *web.Network, err error) {
	var conn Connection
	if conn, err = e.connFactory.CreateConn(); err != nil {
		return
	}
	defer conn.Close()
	var data types.NetworkResource
	if data, err = conn.NetworkInspect(ctx, string(id), types.NetworkInspectOptions{}); err == nil {
		net = &web.Network{ID: id, Name: data.Name}
	}
	return
}
