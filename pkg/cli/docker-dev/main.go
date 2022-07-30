package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/adirelle/docker-graph/pkg/lib/docker"
	"github.com/docker/docker/client"
	"github.com/thejerf/suture/v4"
)

type subscriber struct {
	*json.Encoder
}

func main() {
	spv := suture.NewSimple("docker-graph")

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)

	connFactory := docker.MakeBasicConnectionFactory(client.FromEnv)

	svc := docker.NewEventSource(connFactory)
	spv.Add(svc)

	go func() {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		done := svc.Subscribe(subscriber{enc})

		defer done()
		<-ctx.Done()
	}()

	if err := spv.Serve(ctx); err != nil {
		log.Fatalf("Exiting: %s", err)
	}
}

func (s subscriber) Receive(e docker.Event, ctx context.Context) {
	s.Encode(e)
}
