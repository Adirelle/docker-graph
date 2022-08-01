//go:build !dev

package main

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	"github.com/adirelle/docker-graph/app"
)

func MountAssets(app *fiber.App) {
	app.Use("/", filesystem.New(filesystem.Config{Root: http.FS(app.Assets), PathPrefix: "/public"}))
}
