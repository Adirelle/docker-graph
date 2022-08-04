package containers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

type (
	ID     string
	Status string

	Container struct {
		ID        ID
		CreatedAt time.Time
		UpdatedAt time.Time
		RemovedAt time.Time
		Name      string
		Image     string
		Status    Status
		Healthy   string   `json:",omitempty"`
		Service   string   `json:",omitempty"`
		Project   *Project `json:",omitempty"`
		Networks  map[string]*Network
		Mounts    []Mount
		Ports     map[string]Port
	}

	Project struct {
		Name       string
		WorkingDir string
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

var (
	_ fmt.Stringer = (*ID)(nil)
	_ fmt.Stringer = (*Status)(nil)
)

func (c *Container) LastUpdateTime() time.Time {
	if !c.RemovedAt.IsZero() {
		return c.RemovedAt
	}
	if !c.UpdatedAt.IsZero() {
		return c.UpdatedAt
	}
	return c.CreatedAt
}

func (c *Container) IsRemoved() bool {
	return !c.RemovedAt.IsZero() || c.Status.IsRemoved()
}

func (c *Container) Update(data types.ContainerJSON, when time.Time) {
	if c.CreatedAt.IsZero() {
		c.Init(data)
	}
	c.Status = Status(data.State.Status)
	if c.Status.IsRunning() && data.State.Health != nil {
		c.Healthy = data.State.Health.Status
	} else {
		c.Healthy = ""
	}
	c.updateNetworks(data.NetworkSettings.Networks)
	c.UpdatedAt = when
}

func (c *Container) Init(container types.ContainerJSON) {
	if when, err := time.Parse(time.RFC3339Nano, container.Created); err == nil {
		c.CreatedAt = when
	} else {
		Log.Warn("invalid created timestamp", "timestamp", container.Created, "error", err)
		c.CreatedAt = time.Now()
	}
	c.Name = container.Name[1:]
	c.Image = container.Config.Image
	if project, ok := container.Config.Labels["com.docker.compose.project"]; ok {
		c.Project = &Project{
			Name:       project,
			WorkingDir: container.Config.Labels["com.docker.compose.project.working_dir"],
		}
	} else {
		c.Project = nil
	}
	c.readMounts(container.Mounts)
	c.readPorts(container.NetworkSettings.Ports)
}

func (c *Container) readMounts(mounts []types.MountPoint) {
	for _, mount := range mounts {
		c.Mounts = append(c.Mounts, Mount{
			Type:        string(mount.Type),
			Name:        mount.Name,
			Source:      mount.Source,
			Destination: mount.Destination,
			ReadWrite:   mount.RW,
		})
	}
}

func (c *Container) readPorts(ports nat.PortMap) {
	c.Ports = make(map[string]Port, len(ports))
	for exposed, value := range ports {
		// According to the package, port should be an array of PortBinding
		// However, it does not seems to complu
		if portBinding, ok := getPortBinding(value); ok && portBinding != nil {
			if portNum, err := strconv.Atoi(portBinding.HostPort); err == nil {
				c.Ports[string(exposed)] = Port{portBinding.HostIP, portNum}
			} else {
				Log.Warn("invalid port number", "port", portBinding.HostPort, "error", err)
			}
		} else if !ok {
			Log.Error("unknown port binding value", "value", value)
		}
	}
}

func getPortBinding(something interface{}) (*nat.PortBinding, bool) {
	switch value := something.(type) {
	case nat.PortBinding:
		return &value, true
	case []nat.PortBinding:
		if len(value) == 0 {
			return nil, true
		}
		return &(value[0]), true
	}
	return nil, false
}

func (c *Container) updateNetworks(networks map[string]*network.EndpointSettings) {
	if c.Networks == nil && len(networks) > 0 {
		c.Networks = make(map[string]*Network, len(networks))
	} else if len(networks) == 0 {
		c.Networks = nil
		return
	}
	for key, netData := range networks {
		dest, found := c.Networks[key]
		if !found {
			dest = &Network{}
			c.Networks[key] = dest
			dest.Name = key
		}
		dest.ID = netData.NetworkID
	}
	for key := range c.Networks {
		if _, found := networks[key]; !found {
			delete(c.Networks, key)
		}
	}
}

func (i ID) String() string {
	return string(i)
}

func (s Status) String() string {
	return string(s)
}

func (s Status) IsRunning() bool {
	return s == "running"
}

func (s Status) IsRemoved() bool {
	return s == "removing"
}
