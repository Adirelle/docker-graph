package main

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/thejerf/suture/v4"
)

type (
	WebServer struct {
		Address string
		*fiber.App
	}
)

var _ suture.Service = (*WebServer)(nil)

func NewWebServer(address string) *WebServer {
	return &WebServer{address, fiber.New()}
}

func (w *WebServer) Serve(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			w.Shutdown()
		case <-done:
		}
	}()
	defer close(done)
	return w.Listen(w.Address)
}
