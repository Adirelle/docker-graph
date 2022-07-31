package events

import (
	"time"
)

type (
	Type       string
	TargetType string

	Event struct {
		TargetType TargetType
		TargetID   string
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

func MakeEvent(
	targetType TargetType,
	targetID string,
	eventType Type,
	when time.Time,
	details any,
) Event {
	return Event{
		TargetType: targetType,
		TargetID:   targetID,
		Type:       eventType,
		Time:       when,
		Details:    details,
	}
}
