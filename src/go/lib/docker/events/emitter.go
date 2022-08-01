package events

import (
	"context"
	"fmt"

	log "github.com/inconshreveable/log15"
	"github.com/thejerf/suture/v4"
)

var (
	Log = log.New()
)

type (
	Emitter struct {
		receivers map[Receiver]bool
		commands  chan emitterCommand

		OnNewReceiver func(Receiver)
	}

	Receiver interface {
		Receive(Event, context.Context)
	}

	emitterCommand interface {
		Execute(*Emitter, context.Context)
	}

	emitCommand struct {
		event Event
	}

	subscribeCommand struct {
		receiver Receiver
	}

	unsubscribeCommand struct {
		receiver Receiver
	}
)

var (
	_ suture.Service = (*Emitter)(nil)
	_ fmt.GoStringer = (*Emitter)(nil)

	_ emitterCommand = (*emitCommand)(nil)
	_ emitterCommand = (*subscribeCommand)(nil)
	_ emitterCommand = (*unsubscribeCommand)(nil)
)

func NewEmitter() *Emitter {
	return &Emitter{
		receivers: make(map[Receiver]bool),
		commands:  make(chan emitterCommand, 50),
	}
}

func (e *Emitter) GoString() string {
	return fmt.Sprintf("events.Emitter(%d, %d/%d)", len(e.receivers), len(e.commands), cap(e.commands))
}

func (e *Emitter) Serve(ctx context.Context) error {
	for {
		select {
		case cmd := <-e.commands:
			cmd.Execute(e, ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (e *Emitter) Subscribe(receiver Receiver) func() {
	e.commands <- subscribeCommand{receiver}
	return func() {
		e.commands <- unsubscribeCommand{receiver}
	}
}

func (e *Emitter) Emit(event Event) {
	e.commands <- emitCommand{event}
}

func (c emitCommand) Execute(e *Emitter, ctx context.Context) {
	Log.Debug("emitting event", log.Ctx{
		"target_type": c.event.TargetType,
		"target_id":   c.event.TargetID,
		"event_time":  c.event.Time,
		"event_type":  c.event.Type,
		"details":     c.event.Details,
	})
	for receiver := range e.receivers {
		go receiver.Receive(c.event, ctx)
	}
}

func (c subscribeCommand) Execute(e *Emitter, ctx context.Context) {
	if _, found := e.receivers[c.receiver]; found {
		return
	}
	e.receivers[c.receiver] = true
	if e.OnNewReceiver != nil {
		e.OnNewReceiver(c.receiver)
	}
}

func (c unsubscribeCommand) Execute(e *Emitter, ctx context.Context) {
	delete(e.receivers, c.receiver)
}
