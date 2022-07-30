//go:build dev

package main

import (
	"context"
	"flag"
	"log"
	"net/netip"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/thejerf/suture/v4"

	"github.com/evanw/esbuild/pkg/api"
)

type (
	ESBuilderServer struct {
		addr      netip.AddrPort
		assetPath string
	}
)

var (
	proxyAddr string

	esbuildOptions = api.BuildOptions{
		Bundle:   true,
		Platform: api.PlatformBrowser,
		Format:   api.FormatIIFE,
		Color:    api.ColorAlways,
		LogLimit: 0,
		Outfile:  "public/script.js",
	}

	_ suture.Service = (*ESBuilderServer)(nil)
)

func init() {
	flag.StringVar(&proxyAddr, "internalProxy", "127.0.0.150:38525", "Internal port used for Javascript server")
}

func MakeAssetHandler() (handler fiber.Handler, err error) {
	server := &ESBuilderServer{}

	if server.addr, err = netip.ParseAddrPort(proxyAddr); err != nil {
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	appPath := filepath.Join(cwd, "app")

	server.assetPath = filepath.Join(appPath, "public")
	esbuildOptions.AbsWorkingDir = appPath
	esbuildOptions.EntryPoints = []string{filepath.Join(appPath, "src/index.ts")}

	handler = proxy.Balancer(proxy.Config{Servers: []string{proxyAddr}})

	switch {
	case Debug:
		esbuildOptions.LogLevel = api.LogLevelDebug
	case Quiet:
		esbuildOptions.LogLevel = api.LogLevelWarning
	}

	log.Printf("server=%#v options=%#v", server, esbuildOptions)
	Supervisor.Add(server)

	return
}

func (s *ESBuilderServer) Serve(ctx context.Context) error {
	subCtx, done := context.WithCancel(ctx)
	defer done()

	server, err := api.Serve(
		api.ServeOptions{Port: s.addr.Port(), Host: s.addr.Addr().String(), Servedir: s.assetPath},
		esbuildOptions,
	)
	if err != nil {
		return err
	}

	go func() {
		<-subCtx.Done()
		server.Stop()
	}()

	return server.Wait()
}
