package api

import "github.com/gofiber/fiber/v2"

type (
	API struct {
		eventSource EventSource
		provider    Provider
	}

	Provider interface {
		ContainerProvider
		NetworkProvider
		VolumeProvider
	}
)

func MountAPI(mnt fiber.Router, eventSource EventSource, provider Provider) {
	api := &API{eventSource, provider}
	mnt.Get("/containers", api.listContainers)
	mnt.Get("/containers/:id", api.getContainer)
	mnt.Get("/networks/:id", api.getNetwork)
	mnt.Get("/volumes/:id", api.getVolume)
	mnt.Get("/events", api.streamEvents)
}

func (a *API) streamEvents(ctx *fiber.Ctx) error {
	var lastEventId EventID
	if lastEventIdHeader, found := ctx.GetReqHeaders()["Last-Event-ID"]; found {
		lastEventId = parseEventID(lastEventIdHeader)
	}
	events, stop, err := a.eventSource.Listen(lastEventId)
	if err != nil {
		return err
	}

	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")

	streamer := &Streamer{events: events, stop: stop}
	ctx.Context().SetBodyStreamWriter(streamer.start)

	return nil
}

func (a *API) json(ctx *fiber.Ctx, payload any) error {
	ctx.Set("Content-Type", "application/json")
	ctx.Set("Cache-Control", "no-cache")
	return ctx.JSON(payload)
}

func (a *API) listContainers(ctx *fiber.Ctx) error {
	if ids, err := a.provider.ListContainerIDs(ctx.Context()); err == nil {
		return a.json(ctx, struct{ Containers []ContainerID }{ids})
	} else {
		return err
	}
}

func (a *API) getContainer(ctx *fiber.Ctx) error {
	id := ContainerID(ctx.Params("id"))
	if container, err := a.provider.GetContainer(id, ctx.Context()); err == nil {
		return a.json(ctx, container)
	} else {
		return err
	}
}

func (a *API) getNetwork(ctx *fiber.Ctx) error {
	id := NetworkID(ctx.Params("id"))
	if network, err := a.provider.GetNetwork(id, ctx.Context()); err == nil {
		return a.json(ctx, network)
	} else {
		return err
	}
}

func (a *API) getVolume(ctx *fiber.Ctx) error {
	id := VolumeID(ctx.Params("id"))
	if volume, err := a.provider.GetVolume(id, ctx.Context()); err == nil {
		return a.json(ctx, volume)
	} else {
		return err
	}
}
