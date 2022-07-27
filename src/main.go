package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/thejerf/suture/v4"
)

func main() {
	sup := suture.NewSimple("docker-graph")

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)

	containerIds := make(chan string, 10)
	containerData := make(chan Container, 10)
	go func() {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		for {
			select {
			case c := <-containerData:
				fmt.Print("Container ")
				enc.Encode(c)
			case <-ctx.Done():
			}
		}
	}()

	containerLister := NewDockerClient(&ContainerLister{Output: containerIds})
	sup.Add(containerLister)

	containerInspector := NewDockerClient(ContainerInspector{containerIds, containerData})
	sup.Add(containerInspector)

	webserver := NewWebServer(os.Args[1])
	sup.Add(webserver)

	if err := sup.Serve(ctx); err != nil {
		log.Fatalf("Exiting: %s", err)
	}
}
