package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/adirelle/docker-graph/pkg/lib/api"
	"github.com/adirelle/docker-graph/pkg/lib/docker"
	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/thejerf/suture/v4"
)

type (
	Server struct {
		app     *fiber.App
		address string
	}
)

var (
	_ suture.Service = (*Server)(nil)

	Supervisor = suture.NewSimple("docker-graph")
	Debug      = false
	Verbose    = false
	Quiet      = false
)

func main() {
	flag.BoolVar(&Debug, "debug", false, "Enable ")
	flag.BoolVar(&Verbose, "verbose", false, "Be more verbose")
	flag.BoolVar(&Quiet, "quiet", false, "Disable all output messages but warnings and errors")
	httpAddr := flag.String("bind", ":8080", "Listening address")
	flag.Parse()

	Quiet = Quiet && !Debug
	Verbose = (Verbose || Debug) && !Quiet

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)

	connFactory := docker.MakeBasicConnectionFactory(client.FromEnv)

	eventSource := docker.NewEventSource(connFactory)
	Supervisor.Add(eventSource)

	app := fiber.New(fiber.Config{
		AppName:               "docker-graph",
		ErrorHandler:          handleError,
		DisableStartupMessage: Quiet,
		EnablePrintRoutes:     Debug,
	})
	app.Get("/favicon.ico", favicon.New())
	if Verbose {
		app.Use("/", logger.New())
	}
	if Debug {
		app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	}
	api.MountAPI(app.Group("/api"), eventSource)
	if assetHandler, err := MakeAssetHandler(); err == nil {
		app.Use("/", assetHandler)
	} else {
		log.Fatal(err)
	}

	webserver := NewServer(*httpAddr, app)
	Supervisor.Add(webserver)

	if err := Supervisor.Serve(ctx); err != nil {
		log.Fatalf("Exiting: %s", err)
	}
}

func NewServer(address string, app *fiber.App) (s *Server) {
	return &Server{
		app:     app,
		address: address,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	subCtx, stop := context.WithCancel(ctx)
	defer stop()

	go func() {
		<-subCtx.Done()
		s.app.Shutdown()
	}()

	return s.app.Listen(s.address)
}

func handleError(c *fiber.Ctx, err error) error {
	log.Printf("error: %#v", err)
	fiberError, isFiberError := err.(*fiber.Error)
	switch {
	case isFiberError:
		return c.Status(fiberError.Code).Send([]byte(fiberError.Message))
	case client.IsErrNotFound(err):
		return c.SendStatus(http.StatusNotFound)
	case client.IsErrConnectionFailed(err):
		return c.SendStatus(http.StatusServiceUnavailable)
	default:
		return c.Status(http.StatusInternalServerError).Send([]byte(err.Error()))
	}
}
