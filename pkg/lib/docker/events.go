package docker

import (
	"context"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/thejerf/suture/v4"
)

type (
	EventSource struct {
		connFactory ConnectionFactory

		conn       Connection
		ctx        context.Context
		containers map[ContainerID]*containerWatcher
		receivers  map[Receiver]bool
		commands   chan command
	}

	Receiver interface {
		Receive(Event, context.Context)
	}

	UnsubscribeFunc func()

	containerWatcher struct {
		ID        ContainerID
		LastEvent time.Time
		State     *Container
		Debouncer *time.Timer
	}

	EventType       string
	EventTargetType string

	Event struct {
		TargetType EventTargetType
		TargetID   string
		Type       EventType
		Time       time.Time
		Details    any
	}

	command interface {
		Run(*EventSource)
	}

	subscribeCommand struct {
		receiver Receiver
	}

	unsubscribeCommand struct {
		receiver Receiver
	}
)

const (
	TargetUpdated EventType = "updated"
	TargetRemoved EventType = "removed"

	ContainerTarget EventTargetType = "container"
)

var (
	_ suture.Service = (*EventSource)(nil)

	_ command = (*subscribeCommand)(nil)
	_ command = (*unsubscribeCommand)(nil)
)

func NewEventSource(connFactory ConnectionFactory) *EventSource {
	return &EventSource{
		connFactory: connFactory,
		containers:  make(map[ContainerID]*containerWatcher),
		receivers:   make(map[Receiver]bool),
		commands:    make(chan command, 10),
	}
}

func (s *EventSource) Serve(ctx context.Context) (err error) {
	s.ctx = ctx
	s.conn, err = s.connFactory.CreateConn()
	if err != nil {
		return err
	}
	defer s.conn.Close()

	eventC, errC := s.conn.Events(ctx, types.EventsOptions{Since: "5000h"})

	for {
		select {
		case cmd := <-s.commands:
			cmd.Run(s)
		case msg := <-eventC:
			s.process(msg)
		case err = <-errC:
			return
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *EventSource) Subscribe(receiver Receiver) UnsubscribeFunc {
	s.commands <- subscribeCommand{receiver}
	return func() {
		s.commands <- unsubscribeCommand{receiver}
	}
}

func (s *EventSource) process(msg events.Message) {
	when := time.Unix(0, msg.TimeNano)
	switch msg.Type {
	case "container":
		switch msg.Action {
		case "create", "start", "resize", "kill", "die":
			s.updateContainer(ContainerID(msg.ID), when)
		case "destroy":
			s.removeContainer(ContainerID(msg.ID), when)
		case "attach":
			// Silently discard
		default:
			log.Printf("ignored container event: %s", msg.Action)
		}
	case "network":
		switch msg.Action {
		case "connect", "disconnect":
			s.updateContainer(ContainerID(msg.Actor.Attributes["container"]), when)
		default:
			log.Printf("ignored network event: %s", msg.Action)
		}
	default:
		log.Printf("ignored %s event: %s", msg.Type, msg.Action)
	}
}

func (s *EventSource) updateContainer(id ContainerID, when time.Time) {
	ctn, found := s.containers[id]
	if !found {
		ctn = &containerWatcher{ID: id}
		ctn.Debouncer = time.AfterFunc(time.Second, func() {
			s.doUpdateContainer(ctn)
		})
		s.containers[id] = ctn
		// log.Printf("new container: %s", id)
	}
	ctn.LastEvent = when
	ctn.Debouncer.Reset(500 * time.Millisecond)
}

func (s *EventSource) doUpdateContainer(ctn *containerWatcher) {
	data, err := s.conn.ContainerInspect(s.ctx, string(ctn.ID))
	if err != nil {
		log.Println(err)
		return
	}
	// repr, _ := json.MarshalIndent(data, "", "  ")
	// log.Printf("data: %s", repr)

	if ctn.State == nil {
		ctn.State = NewContainer(&data)
	} else {
		ctn.State.Update(&data)
	}
	s.emit(ctn.MakeEvent())
}

func (s *EventSource) removeContainer(id ContainerID, when time.Time) {
	ctn, found := s.containers[id]
	if !found {
		return
	}
	if !ctn.Debouncer.Stop() {
		<-ctn.Debouncer.C
	}
	delete(s.containers, id)
	ctn.State = nil
	s.emit(ctn.MakeEvent())
}

func (s *EventSource) emit(event Event) {
	for receiver := range s.receivers {
		go receiver.Receive(event, s.ctx)
	}
}

func (c subscribeCommand) Run(p *EventSource) {
	if _, found := p.receivers[c.receiver]; found {
		return
	}
	for _, container := range p.containers {
		c.receiver.Receive(container.MakeEvent(), p.ctx)
	}
	p.receivers[c.receiver] = true
}

func (c unsubscribeCommand) Run(p *EventSource) {
	delete(p.receivers, c.receiver)
}

func (w containerWatcher) MakeEvent() (e Event) {
	e.TargetID = string(w.ID)
	e.TargetType = ContainerTarget
	e.Time = w.LastEvent
	if w.State != nil {
		e.Type = TargetUpdated
		e.Details = w.State
	} else {
		e.Type = TargetRemoved
	}
	return
}
