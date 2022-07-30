package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/adirelle/docker-graph/pkg/lib/docker"
	"github.com/docker/docker/client"
	"github.com/thejerf/suture/v4"
)

func main() {
	spv := suture.NewSimple("docker-graph")

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)

	connFactory := docker.NewConnectionPool(docker.MakeBasicConnectionFactory(client.FromEnv))

	events := docker.NewStreamFactory(connFactory)
	spv.Add(events)

	if err := spv.Serve(ctx); err != nil {
		log.Fatalf("Exiting: %s", err)
	}
}
