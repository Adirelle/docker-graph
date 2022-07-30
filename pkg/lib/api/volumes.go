package api

import "context"

type (
	VolumeID string

	Volume struct {
		ID   VolumeID `json:"id"`
		Name string   `json:"name"`
	}

	VolumeProvider interface {
		GetVolume(VolumeID, context.Context) (*Volume, error)
	}
)
