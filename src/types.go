package main

import "github.com/docker/docker/api/types"

type (
	Container struct {
		ID           string
		Name         string
		Status       string
		Dependencies []string
		Mounts       []string
		Networks     []string
	}

	Volume struct {
		ID   string
		Name string
	}

	Network struct {
		ID   string
		Name string
	}
)

func ConvertContainer(in types.ContainerJSON) (out Container) {
	out.ID = in.ID
	out.Name = in.Name
	out.Status = in.State.Status
	for _, mount := range in.Mounts {
		out.Mounts = append(out.Mounts, mount.Name)
	}
	for _, network := range in.NetworkSettings.Networks {
		if network.NetworkID != "" {
			out.Networks = append(out.Networks, network.NetworkID)
		}
	}
	return
}
