//go:build !dev

package main

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	assets "github.com/adirelle/docker-graph"
)

func MountAssets(app *fiber.App) {
	app.Use("/", filesystem.New(filesystem.Config{Root: http.FS(assets.Assets), PathPrefix: "/public"}))
}
