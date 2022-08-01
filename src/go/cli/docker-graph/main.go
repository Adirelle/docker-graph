package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/adirelle/docker-graph/src/go/lib/api"
	"github.com/adirelle/docker-graph/src/go/lib/docker/connections"
	"github.com/adirelle/docker-graph/src/go/lib/docker/containers"
	"github.com/adirelle/docker-graph/src/go/lib/docker/events"
	"github.com/adirelle/docker-graph/src/go/lib/docker/listeners"
	"github.com/docker/docker/client"
	log "github.com/inconshreveable/log15"
	"github.com/thejerf/suture/v4"
)

func main() {
	flag.Parse()

	SetupLogging()

	spv := suture.New("docker-graph", suture.Spec{
		EventHook: func(ev suture.Event) {
			Log.Error(ev.String(), "type", ev.Type(), "context", log.Ctx(ev.Map()))
		},
	})

	eventEmitter := events.NewEmitter()
	spv.Add(eventEmitter)

	connFactory := connections.MakeBasicFactory(client.FromEnv)

	containerRepo := containers.NewRepository(eventEmitter, connFactory)
	spv.Add(containerRepo)

	listener := listeners.NewListener(connFactory, containerRepo)
	spv.Add(listener)

	webserver := NewWebServer()
	spv.Add(webserver)

	eventAPI := api.NewAPI(eventEmitter)
	eventAPI.MountInto(webserver.App.Group("/api"))

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt, syscall.SIGHUP)
	if err := spv.Serve(ctx); err != nil {
		Log.Crit("Exiting: %s", err)
	}
}
