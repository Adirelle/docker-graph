//go:build !dev

package main

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	"github.com/adirelle/docker-graph/app"
)

func MakeAssetHandler() (fiber.Handler, error) {
	return filesystem.New(filesystem.Config{Root: http.FS(app.Assets), PathPrefix: "/public"}), nil
}
