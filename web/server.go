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
		events     EventSource
		containers ContainerProvider
		networks   NetworkProvider
		address    string
		devMode    bool
	}
)

var (
	_ suture.Service = (*Server)(nil)

	//go:embed static
	assets embed.FS
)

func NewServer(address string, devMode bool, events EventSource, containers ContainerProvider, networks NetworkProvider) *Server {
	return &Server{
		events:     events,
		containers: containers,
		networks:   networks,
		devMode:    devMode,
		address:    address,
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

	app.Get("/api/networks", s.listNetworks)
	app.Get("/api/networks/:id", s.getNetwork)

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
	events, stop, err := s.events.Listen(lastEventId)
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
	if ids, err := s.containers.ListContainerIDs(ctx.Context()); err == nil {
		return s.json(ctx, struct{ Containers []ContainerID }{ids})
	} else {
		return err
	}
}

func (s *Server) getContainer(ctx *fiber.Ctx) error {
	id := ContainerID(ctx.Params("id"))
	if container, err := s.containers.GetContainer(id, ctx.Context()); err == nil {
		return s.json(ctx, container)
	} else {
		return err
	}
}

func (s *Server) listNetworks(ctx *fiber.Ctx) error {
	if ids, err := s.networks.ListNetworkIDs(ctx.Context()); err == nil {
		return s.json(ctx, struct{ Networks []NetworkID }{ids})
	} else {
		return err
	}
}

func (s *Server) getNetwork(ctx *fiber.Ctx) error {
	id := NetworkID(ctx.Params("id"))
	if network, err := s.networks.GetNetwork(id, ctx.Context()); err == nil {
		return s.json(ctx, network)
	} else {
		return err
	}
}
