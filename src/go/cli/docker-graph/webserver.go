package main

import (
	"context"
	"flag"
	"net/http"
	"net/netip"

	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	log "github.com/inconshreveable/log15"
	"github.com/thejerf/suture/v4"
)

type (
	WebServer struct {
		*fiber.App
		log.Logger

		address string
	}

	WebServerBindPort struct {
		netip.AddrPort
	}
)

var (
	_ suture.Service = (*WebServer)(nil)

	bindPort = WebServerBindPort{netip.MustParseAddrPort("127.0.0.1:8080")}
)

func init() {
	flag.Var(&bindPort, "bind", "Listening address")
}

func NewWebServer() (s *WebServer) {
	s = &WebServer{
		address: bindPort.String(),
		Logger:  Log.New("module", "webserver"),
	}

	s.App = fiber.New(fiber.Config{
		AppName:      "docker-graph",
		ErrorHandler: s.handleError,
	})

	s.App.Get("/favicon.ico", favicon.New())
	s.App.Use(s.logRequest)

	return
}

func (s *WebServer) Serve(ctx context.Context) error {
	subCtx, stop := context.WithCancel(ctx)
	defer stop()

	go func() {
		<-subCtx.Done()
		s.App.Shutdown()
	}()

	return s.App.Listen(s.address)
}

func (s *WebServer) handleError(c *fiber.Ctx, err error) error {
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

func (s *WebServer) logRequest(c *fiber.Ctx) (err error) {
	logger := s.Logger.New(
		"method", c.Method()[:],
		"url", c.OriginalURL()[:],
		"proto", c.Protocol()[:],
		"host", c.Hostname()[:],
		"remote", c.IP()[:],
	)
	c.Locals("logger", logger)

	err = c.Next()

	if err != nil {
		s.Logger.Error("request", "error", err)
	} else {
		s.Logger.Info("request", "status", c.Response().StatusCode())
	}

	return
}

func (p *WebServerBindPort) Set(value string) (err error) {
	p.AddrPort, err = netip.ParseAddrPort(value)
	return
}
