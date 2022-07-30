package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/adirelle/docker-graph/pkg/lib/docker"
	"github.com/gofiber/fiber/v2"
)

type (
	API struct {
		source *docker.EventSource
	}

	receiver struct {
		events chan<- docker.Event
	}
)

func MountAPI(mnt fiber.Router, source *docker.EventSource) {
	api := &API{source}
	mnt.Get("/events", api.streamEvents)
}

func (a *API) streamEvents(ctx *fiber.Ctx) error {
	var lastEventTime time.Time
	if lastEventIdHeader := ctx.Get("Last-Event-ID"); lastEventIdHeader != "" {
		if timestamp, err := strconv.ParseInt(lastEventIdHeader, 10, 64); err == nil {
			lastEventTime = time.Unix(0, timestamp)
		} else {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("invalid Last-Event-ID: %s", err))
		}
	}

	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")

	ctx.Context().SetBodyStreamWriter(func(output *bufio.Writer) {
		var err error
		defer func() {
			if err != nil && err != io.EOF {
				log.Println(err)
			}
		}()

		events := make(chan docker.Event, 5)
		defer close(events)

		rcv := receiver{events}
		done := a.source.Subscribe(rcv)
		defer done()

		enc := json.NewEncoder(output)
		for event := range events {
			if event.Time.Before(lastEventTime) {
				continue
			}
			lastEventTime = event.Time
			if _, err = fmt.Fprintf(output, "id:%d\ndata:", event.Time.UnixNano()); err != nil {
				return
			}
			if err = enc.Encode(event); err != nil {
				return
			}
			if _, err = output.WriteString("\n\n"); err != nil {
				return
			}
			if err = output.Flush(); err != nil {
				return
			}
		}
	})

	return nil
}

func (r receiver) Receive(event docker.Event, ctx context.Context) {
	select {
	case r.events <- event:
	case <-ctx.Done():
	}
}
