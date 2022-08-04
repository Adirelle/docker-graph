package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gofiber/fiber/v2"

	log "github.com/inconshreveable/log15"
)

type (
	API struct {
		source EventSource
	}

	EventSource interface {
		Subscribe() (c <-chan Event, cancel func())
	}

	Event interface {
		ID() string
		Data() any
	}
)

func NewAPI(source EventSource) *API {
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

	logger.Debug("starting event stream")

	ctx.Context().SetBodyStreamWriter(func(output *bufio.Writer) {
		var err error
		defer func() {
			if err != nil && err != io.EOF {
				logger.Error("streaming error", err)
			} else {
				logger.Debug("event stream ended")
			}
		}()

		events, done := a.source.Subscribe()
		defer done()

		logger.Debug("waiting for events")
		enc := json.NewEncoder(output)
		for event := range events {
			logger.Debug("sending events", "event", event)
			if err := sendEvent(output, enc, event); err != nil {
				logger.Error("error sending event", "event", event, "error", err)
			} else {
				logger.Debug("sent event", "event", event)
			}
		}
	})

	return nil
}

func sendEvent(output *bufio.Writer, enc *json.Encoder, event Event) (err error) {
	if _, err = fmt.Fprintf(output, "id:%s\ndata:", event.ID()); err != nil {
		return
	}
	if err = enc.Encode(event.Data()); err != nil {
		return
	}
	if _, err = output.WriteString("\n\n"); err != nil {
		return
	}
	return output.Flush()
}
