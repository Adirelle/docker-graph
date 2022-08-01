//go:build dev

package main

import (
	"bytes"
	"flag"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

var (
	devProxy string
)

func init() {
	flag.StringVar(&devProxy, "devProxy", "", "Frontend Proxy URL")
}

func MountAssets(app *fiber.App) {
	if devProxy == "" {
		return
	}
	app.Use("/",
		proxy.Balancer(proxy.Config{
			Servers: []string{devProxy},
			ModifyRequest: func(ctx *fiber.Ctx) error {
				uri := ctx.Request().URI()
				if bytes.Equal(uri.Path(), []byte("/js/index.js")) {
					uri.SetPath("/src/ts/index.ts")
				}
				return nil
			},
		}))
}
