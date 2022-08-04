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
	"github.com/adirelle/docker-graph/src/go/lib/docker/listeners"
	"github.com/adirelle/docker-graph/src/go/lib/logging"
	"github.com/adirelle/docker-graph/src/go/lib/utils"
	"github.com/docker/docker/client"
	log "github.com/inconshreveable/log15"
	"github.com/thejerf/suture/v4"
)

var (
	Log = log.New()
)

func main() {
	webLogger := Log.New(logging.ModuleKey, "webserver")
	utils.Log = Log.New(logging.ModuleKey, "dispatcher")

	dockerLogger := Log.New(logging.ModuleKey, "docker")
	connections.Log = dockerLogger.New(logging.ModuleKey, "connections")
	containers.Log = dockerLogger.New(logging.ModuleKey, "containers")
	listeners.Log = dockerLogger.New(logging.ModuleKey, "listeners")

	logConfig := logging.Config{
		Modules:     logging.ModuleLevels{logging.MainModule: log.LvlWarn},
		StderrLevel: logging.Level(log.LvlDebug),
	}
	logConfig.SetupFlags()

	flag.Parse()

	logConfig.Apply(Log)

	spv := suture.New("docker-graph", suture.Spec{
		EventHook: func(ev suture.Event) {
			Log.Error(ev.String(), "type", ev.Type(), "context", log.Ctx(ev.Map()))
		},
	})

	dispatcher := utils.NewDispatcher[api.Event]()
	spv.Add(dispatcher)

	connFactory := connections.MakeBasicFactory(client.FromEnv)

	containerRepo := containers.NewRepository(dispatcher, connFactory)
	spv.Add(containerRepo)

	listener := listeners.NewListener(connFactory, containerRepo)
	spv.Add(listener)

	webserver := NewWebServer(webLogger)
	spv.Add(webserver)

	eventAPI := api.NewAPI(dispatcher)
	eventAPI.MountInto(webserver.App.Group("/api"))

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt, syscall.SIGHUP)
	if err := spv.Serve(ctx); err != nil {
		Log.Crit("Exiting: %s", err)
	}
}
