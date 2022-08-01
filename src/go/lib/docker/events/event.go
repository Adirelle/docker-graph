package events

import (
	"fmt"
	"time"
)

type (
	Type       string
	TargetType string

	ID interface {
		fmt.Stringer
	}

	Event struct {
		TargetType TargetType
		TargetID   ID
		Type       Type
		Time       time.Time
		Details    any
	}
)

const (
	TargetUpdated Type = "updated"
	TargetRemoved Type = "removed"

	ContainerTarget TargetType = "container"
)

func MakeContainerUpdatedEvent(id ID, when time.Time, details any) (e Event) {
	e.TargetType = ContainerTarget
	e.TargetID = id
	e.Type = TargetUpdated
	e.Time = when
	e.Details = details
	return
}

func MakeContainerRemovedEvent(id ID, when time.Time) (e Event) {
	e.TargetType = ContainerTarget
	e.TargetID = id
	e.Type = TargetRemoved
	e.Time = when
	return
}
