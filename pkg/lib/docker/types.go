package docker

import (
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

type (
	ContainerID     string
	ContainerStatus string

	Container struct {
		ID       ContainerID
		Name     string
		Image    string
		Status   ContainerStatus
		Healthy  string `json:",omitempty"`
		Service  string `json:",omitempty"`
		Project  string `json:",omitempty"`
		Networks map[string]*Network
		Mounts   []Mount
		Ports    map[string]Port
	}

	Network struct {
		ID   string
		Name string
	}

	Mount struct {
		Name        string
		Type        string
		Source      string
		Destination string
		ReadWrite   bool
	}

	Port struct {
		HostIp   string
		HostPort int
	}
)

func NewContainer(container *types.ContainerJSON) *Container {
	c := &Container{
		ID:      ContainerID(container.ID),
		Name:    container.Name[1:],
		Image:   container.Config.Image,
		Project: container.Config.Labels["com.docker.compose.project"],
		Service: container.Config.Labels["com.docker.compose.service"],
	}
	c.readMounts(container.Config.Labels["com.docker.compose.project.working_dir"], container.Mounts)
	c.readPorts(container.NetworkSettings.Ports)
	c.Update(container)
	return c
}

func (c *Container) readMounts(baseDir string, mounts []types.MountPoint) {
	if baseDir != "" {
		baseDir += "/"
	}
	for _, mount := range mounts {
		src := mount.Source
		if mount.Type == "bind" && strings.HasPrefix(src, baseDir) {
			src = src[len(baseDir):]
		}
		c.Mounts = append(c.Mounts, Mount{
			Type:        string(mount.Type),
			Name:        mount.Name,
			Source:      src,
			Destination: mount.Destination,
			ReadWrite:   mount.RW,
		})
	}
}

func (c *Container) readPorts(ports nat.PortMap) {
	c.Ports = make(map[string]Port, len(ports))
	for key, port := range ports {
		portNum, _ := strconv.Atoi(port[0].HostPort)
		c.Ports[string(key)] = Port{port[0].HostIP, portNum}
	}
}

func (c *Container) Update(container *types.ContainerJSON) {
	c.Status = ContainerStatus(container.State.Status)
	if container.State.Health != nil {
		c.Healthy = container.State.Health.Status
	} else {
		c.Healthy = ""
	}
	c.updateNetworks(container.NetworkSettings.Networks)
}

func (c *Container) updateNetworks(networks map[string]*network.EndpointSettings) {
	if c.Networks == nil {
		c.Networks = make(map[string]*Network, len(networks))
	}
	for key, netData := range networks {
		dest, found := c.Networks[key]
		if !found {
			dest = &Network{}
			c.Networks[key] = dest
		}
		dest.ID = netData.NetworkID
		dest.Name = append(netData.Aliases, key)[0]
	}
	for key := range c.Networks {
		if _, found := networks[key]; !found {
			delete(c.Networks, key)
		}
	}
}
