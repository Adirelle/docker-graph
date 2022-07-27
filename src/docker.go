package main

import (
	"context"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/thejerf/suture/v4"
)

type (
	DockerClient struct {
		client.Opt
		DockerService
	}

	DockerService interface {
		Serve(context.Context, *client.Client) error
	}

	ContainerLister struct {
		Output   chan<- string
		Primed   bool
		LastTime time.Time
	}

	ContainerInspector struct {
		Input  <-chan string
		Output chan<- Container
	}
)

var (
	_ suture.Service = (*DockerClient)(nil)

	containerEventFilter = filters.NewArgs(filters.Arg("Type", "container"))

	eventWhiteList = map[string]bool{
		"create":      true,
		"start":       true,
		"die":         true,
		"destroy":     true,
		"connect":     true,
		"disconnnect": true,
	}
)

func NewDockerClient(svc DockerService) *DockerClient {
	return &DockerClient{client.FromEnv, svc}
}

func (c *DockerClient) Serve(ctx context.Context) error {
	client, err := client.NewClientWithOpts(c.Opt)
	if err != nil {
		return err
	}
	defer client.Close()

	return c.DockerService.Serve(ctx, client)
}

func (l *ContainerLister) Serve(ctx context.Context, client *client.Client) error {
	if !l.Primed {
		containers, err := client.ContainerList(ctx, types.ContainerListOptions{})
		if err != nil {
			return err
		}
		l.Primed = true
		l.LastTime = time.Now()
		for _, cont := range containers {
			l.Output <- cont.ID
		}
	}

	events, errC := client.Events(
		ctx,
		types.EventsOptions{Since: l.LastTime.Format(time.RFC3339), Filters: containerEventFilter},
	)
	for {
		select {
		case msg := <-events:
			l.LastTime = time.UnixMilli(msg.Time)
			if accept, found := eventWhiteList[msg.Action]; accept && found {
				if msg.Type == "container" {
					l.Output <- msg.ID
				} else if msg.Type == "network" {
					l.Output <- msg.Actor.Attributes["container"]
				}
			} else {
				log.Printf("event ignored: %#v", msg)
			}
		case err := <-errC:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (i ContainerInspector) Serve(ctx context.Context, dockerClient *client.Client) error {
	for {
		select {
		case id := <-i.Input:
			log.Printf("inspecting %s", id)
			container, err := dockerClient.ContainerInspect(ctx, id)
			if err == nil {
				select {
				case i.Output <- ConvertContainer(container):
				case <-ctx.Done():
					return ctx.Err()
				}
			} else if client.IsErrNotFound(err) {
				// NOOP
			} else {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
