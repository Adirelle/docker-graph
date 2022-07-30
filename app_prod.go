//go:build !dev

package main

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:generate cd app; bun run build
//go:embed app/public
var assets embed.FS

func MakeAssetHandler() (fiber.Handler, error) {
	return filesystem.New(filesystem.Config{Root: http.FS(assets), PathPrefix: "/app/public"}), nil
}
