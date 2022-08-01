package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/adirelle/docker-graph/src/go/lib/docker/events"
	"github.com/gofiber/fiber/v2"

	log "github.com/inconshreveable/log15"
)

type (
	API struct {
		source *events.Emitter
	}

	receiver struct {
		events chan<- events.Event
	}
)

func NewAPI(source *events.Emitter) *API {
	return &API{source}
}

func (a *API) MountInto(mnt fiber.Router) {
	mnt.Get("/events", a.streamEvents)
}

func (a *API) streamEvents(ctx *fiber.Ctx) error {
	logger := ctx.Locals("logger").(log.Logger)

	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")

	ctx.Context().SetBodyStreamWriter(func(output *bufio.Writer) {
		var err error
		defer func() {
			if err != nil && err != io.EOF {
				logger.Error("streaming error", err)
			}
		}()

		events := make(chan events.Event, 5)
		defer close(events)

		rcv := receiver{events}
		done := a.source.Subscribe(rcv)
		defer done()

		enc := json.NewEncoder(output)
		for event := range events {
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
			logger.Debug("sent event", "event", event)
		}
	})

	return nil
}

func (r receiver) Receive(event events.Event, ctx context.Context) {
	select {
	case r.events <- event:
	case <-ctx.Done():
	}
}
