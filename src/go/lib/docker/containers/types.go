package containers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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
		Healthy   string `json:",omitempty"`
		Service   string `json:",omitempty"`
		Project   string `json:",omitempty"`
		Networks  map[string]*Network
		Mounts    []Mount
		Ports     map[string]Port
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

func (c *Container) Update(data types.ContainerJSON, when time.Time) (changed bool) {
	if c.CreatedAt.IsZero() {
		c.Init(data)
		changed = true
	}
	if status := Status(data.State.Status); c.Status != status {
		c.Status = status
		changed = true
		log.Printf("status=%s", status)
		if status.IsRemoved() {
			c.RemovedAt = when
		}
	}
	healthy := ""
	if c.Status.IsRunning() && data.State.Health != nil {
		healthy = data.State.Health.Status
	}
	if c.Healthy != healthy {
		c.Healthy = healthy
		changed = true
	}
	if c.updateNetworks(data.NetworkSettings.Networks) {
		changed = true
	}
	if changed {
		c.UpdatedAt = when
	}
	return changed
}

func (c *Container) Init(container types.ContainerJSON) {
	if when, err := time.Parse(time.RFC3339Nano, container.Created); err == nil {
		c.CreatedAt = when
	} else {
		log.Printf("invalid created timestamp: %s", container.Created)
		c.CreatedAt = time.Now()
	}
	c.Name = container.Name[1:]
	c.Image = container.Config.Image
	c.Project = container.Config.Labels["com.docker.compose.project"]
	c.Service = container.Config.Labels["com.docker.compose.service"]
	c.readMounts(container.Config.Labels["com.docker.compose.project.working_dir"], container.Mounts)
	c.readPorts(container.NetworkSettings.Ports)
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
	for exposed, value := range ports {
		// According to the package, port should be an array of PortBinding
		// However, it does not seems to complu
		if portBinding, ok := getPortBinding(value); ok && portBinding != nil {
			if portNum, err := strconv.Atoi(portBinding.HostPort); err == nil {
				c.Ports[string(exposed)] = Port{portBinding.HostIP, portNum}
			} else {
				log.Printf("invalid port number: `%s`: %s", portBinding.HostPort, err)
			}
		} else if !ok {
			log.Printf("unknown port binding value: %#v", value)
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

func (c *Container) updateNetworks(networks map[string]*network.EndpointSettings) (changed bool) {
	if c.Networks == nil && len(networks) > 0 {
		c.Networks = make(map[string]*Network, len(networks))
		changed = true
	} else if len(networks) == 0 && c.Networks != nil {
		c.Networks = nil
		return true
	}
	for key, netData := range networks {
		dest, found := c.Networks[key]
		if !found {
			dest = &Network{}
			c.Networks[key] = dest
			dest.Name = key
			changed = true
		}
		if dest.ID != netData.NetworkID {
			dest.ID = netData.NetworkID
			changed = true
		}
	}
	for key := range c.Networks {
		if _, found := networks[key]; !found {
			delete(c.Networks, key)
			changed = true
		}
	}

	return
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
