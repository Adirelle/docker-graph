package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/adirelle/docker-graph/docker"
	"github.com/adirelle/docker-graph/web"
	"github.com/docker/docker/client"
	"github.com/thejerf/suture/v4"
)

func main() {
	devMode := false
	httpAddr := ":80"
	flag.BoolVar(&devMode, "dev", false, "Enable development mode")
	flag.StringVar(&httpAddr, "bind", ":8080", "Listening address")
	flag.Parse()

	sup := suture.NewSimple("docker-graph")

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)

	connFactory := docker.NewConnectionPool(docker.MakeBasicConnectionFactory(client.FromEnv))
	endpoint := docker.NewEndpoint(connFactory)

	events := docker.NewStreamFactory(connFactory)
	sup.Add(events)

	webserver := web.NewServer(httpAddr, devMode, events, endpoint, endpoint)
	sup.Add(webserver)

	if err := sup.Serve(ctx); err != nil {
		log.Fatalf("Exiting: %s", err)
	}
}
