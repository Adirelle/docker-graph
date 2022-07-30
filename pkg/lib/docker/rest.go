package docker

import (
	"context"
	"strings"

	"github.com/adirelle/docker-graph/pkg/lib/api"
	"github.com/docker/docker/api/types"
)

type (
	RESTEndpoint struct {
		connFactory ConnectionFactory
	}
)

var (
	_ api.ContainerProvider = (*RESTEndpoint)(nil)
	_ api.NetworkProvider   = (*RESTEndpoint)(nil)
	_ api.VolumeProvider    = (*RESTEndpoint)(nil)
)

func NewRESTEndpoint(connFactory ConnectionFactory) *RESTEndpoint {
	return &RESTEndpoint{connFactory}
}

func (e *RESTEndpoint) ListContainerIDs(ctx context.Context) (ids []api.ContainerID, err error) {
	var conn Connection
	if conn, err = e.connFactory.CreateConn(); err != nil {
		return
	}
	defer conn.Close()
	var containers []types.Container
	if containers, err = conn.ContainerList(ctx, types.ContainerListOptions{}); err == nil {
		for _, container := range containers {
			ids = append(ids, api.ContainerID(container.ID))
		}
	}
	return
}

func (e *RESTEndpoint) GetContainer(id api.ContainerID, ctx context.Context) (ctn *api.Container, err error) {
	var conn Connection
	if conn, err = e.connFactory.CreateConn(); err != nil {
		return
	}
	defer conn.Close()
	var data types.ContainerJSON
	if data, err = conn.ContainerInspect(ctx, string(id)); err == nil {
		ctn = &api.Container{
			ID:      id,
			Name:    data.Name,
			Status:  data.State.Status,
			Project: data.Config.Labels["com.docker.compose.project"],
			Service: data.Config.Labels["com.docker.compose.service"],
		}

		baseDir := ""
		if workDir, found := data.Config.Labels["com.docker.compose.project.working_dir"]; found {
			baseDir = workDir + "/"
		}

		for _, net := range data.NetworkSettings.Networks {
			ctn.Networks = append(ctn.Networks, api.NetworkID(net.NetworkID))
		}
		for _, mnt := range data.Mounts {
			apiMount := api.Mount{
				Type:        string(mnt.Type),
				Name:        mnt.Name,
				Source:      mnt.Source,
				Destination: mnt.Destination,
				ReadWrite:   mnt.RW,
			}
			if apiMount.Type == "bind" && baseDir != "" && strings.HasPrefix(apiMount.Source, baseDir) {
				apiMount.Source = "./" + apiMount.Source[len(baseDir):]
			}
			ctn.Mounts = append(ctn.Mounts, apiMount)
		}
	}
	return
}

func (e *RESTEndpoint) GetNetwork(id api.NetworkID, ctx context.Context) (net *api.Network, err error) {
	var conn Connection
	if conn, err = e.connFactory.CreateConn(); err != nil {
		return
	}
	defer conn.Close()
	var data types.NetworkResource
	if data, err = conn.NetworkInspect(ctx, string(id), types.NetworkInspectOptions{}); err == nil {
		net = &api.Network{ID: id, Name: data.Name}
	}
	return
}

func (e *RESTEndpoint) GetVolume(id api.VolumeID, ctx context.Context) (vol *api.Volume, err error) {
	var conn Connection
	if conn, err = e.connFactory.CreateConn(); err != nil {
		return
	}
	defer conn.Close()
	var data types.Volume
	if data, err = conn.VolumeInspect(ctx, string(id)); err == nil {
		vol = &api.Volume{ID: id, Name: data.Name}
	}
	return
}
