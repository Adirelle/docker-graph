package containers

import (
	"time"

	"github.com/adirelle/docker-graph/src/go/lib/api"
)

type (
	ContainerUpdated struct {
		when time.Time
		data *Container
	}

	ContainerRemoved struct {
		when time.Time
		id   string
	}

	eventDTO struct {
		TargetType string
		TargetID   string
		Type       string
		Time       time.Time
		Details    any
	}
)

var (
	_ api.Event = (*ContainerUpdated)(nil)
	_ api.Event = (*ContainerRemoved)(nil)
)

const (
	IDFormat = time.RFC3339Nano
)

func (c *ContainerUpdated) ID() string {
	return c.when.Format(IDFormat)
}

func (c *ContainerUpdated) Data() any {
	return eventDTO{
		TargetType: "container",
		TargetID:   string(c.data.ID),
		Type:       "updated",
		Time:       c.when,
		Details:    c.data,
	}
}

func (c *ContainerRemoved) ID() string {
	return c.when.Format(IDFormat)
}

func (c *ContainerRemoved) Data() any {
	return eventDTO{
		TargetType: "container",
		TargetID:   c.id,
		Type:       "removed",
		Time:       c.when,
	}
}
