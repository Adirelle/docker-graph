//go:build !dev

package main

import (
	"net/http"

	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	assets "github.com/adirelle/docker-graph"
)

func (s *WebServer) MountAssets() {
	s.App.Use("/",
		compress.New(),
		etag.New(),
		filesystem.New(filesystem.Config{
			Root:       http.FS(assets.Assets),
			PathPrefix: "/public",
		}),
	)
}
