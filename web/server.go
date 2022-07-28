package web

import (
	"context"
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/thejerf/suture/v4"
)

type (
	Server struct {
		provider    Provider
		eventSource EventSource
		address     string
		devMode     bool
	}

	Provider interface {
		ContainerProvider
		NetworkProvider
		VolumeProvider
	}
)

var (
	_ suture.Service = (*Server)(nil)

	//go:embed static
	assets embed.FS
)

func NewServer(address string, devMode bool, eventSource EventSource, provider Provider) *Server {
	return &Server{
		provider:    provider,
		eventSource: eventSource,
		devMode:     devMode,
		address:     address,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	app := s.createApp()
	subCtx, stop := context.WithCancel(ctx)
	defer stop()

	go func() {
		<-subCtx.Done()
		app.Shutdown()
	}()

	return app.Listen(s.address)
}

func (s *Server) createApp() (app *fiber.App) {
	app = fiber.New()

	app.Use(
		recover.New(recover.Config{EnableStackTrace: s.devMode}),
		favicon.New(),
		logger.New(),
	)

	app.Get("/api/containers", s.listContainers)
	app.Get("/api/containers/:id", s.getContainer)
	app.Get("/api/networks/:id", s.getNetwork)
	app.Get("/api/volumes/:id", s.getVolume)
	app.Get("/api/events", s.streamEvents)

	if s.devMode {
		app.Static("/", "web/static")
	} else {
		app.Use("/", filesystem.New(filesystem.Config{Root: http.FS(assets), PathPrefix: "/static"}))
	}

	return
}

func (s *Server) streamEvents(ctx *fiber.Ctx) error {
	var lastEventId EventID
	if lastEventIdHeader, found := ctx.GetReqHeaders()["Last-Event-ID"]; found {
		lastEventId = parseEventID(lastEventIdHeader)
	}
	events, stop, err := s.eventSource.Listen(lastEventId)
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

func (s *Server) json(ctx *fiber.Ctx, payload any) error {
	ctx.Set("Content-Type", "application/json")
	ctx.Set("Cache-Control", "no-cache")
	return ctx.JSON(payload)
}

func (s *Server) listContainers(ctx *fiber.Ctx) error {
	if ids, err := s.provider.ListContainerIDs(ctx.Context()); err == nil {
		return s.json(ctx, struct{ Containers []ContainerID }{ids})
	} else {
		return err
	}
}

func (s *Server) getContainer(ctx *fiber.Ctx) error {
	id := ContainerID(ctx.Params("id"))
	if container, err := s.provider.GetContainer(id, ctx.Context()); err == nil {
		return s.json(ctx, container)
	} else {
		return err
	}
}

func (s *Server) getNetwork(ctx *fiber.Ctx) error {
	id := NetworkID(ctx.Params("id"))
	if network, err := s.provider.GetNetwork(id, ctx.Context()); err == nil {
		return s.json(ctx, network)
	} else {
		return err
	}
}

func (s *Server) getVolume(ctx *fiber.Ctx) error {
	id := VolumeID(ctx.Params("id"))
	if volume, err := s.provider.GetVolume(id, ctx.Context()); err == nil {
		return s.json(ctx, volume)
	} else {
		return err
	}
}
