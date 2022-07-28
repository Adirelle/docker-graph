package web

import "context"

type (
	VolumeID string

	Volume struct {
		ID   VolumeID
		Name string
	}

	VolumeProvider interface {
		GetVolume(VolumeID, context.Context) (*Volume, error)
	}
)
