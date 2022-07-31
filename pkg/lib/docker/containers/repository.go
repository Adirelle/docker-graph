package containers

import (
	"context"
	"log"
	"time"

	"github.com/adirelle/docker-graph/pkg/lib/docker/connections"
	myEvents "github.com/adirelle/docker-graph/pkg/lib/docker/events"
	"github.com/docker/docker/api/types/events"
	"github.com/thejerf/suture/v4"
)

type (
	Repository struct {
		Emitter     *myEvents.Emitter
		ConnFactory connections.Factory

		conn       connections.Connection
		commands   chan repoCommand
		containers map[ID]*Container
	}

	repoCommand interface {
		Execute(*Repository, context.Context) error
	}

	processMessageCmd struct {
		msg events.Message
	}

	inspectCmd struct {
		id   ID
		when time.Time
	}

	removeCmd struct {
		id   ID
		when time.Time
	}
)

var (
	_ suture.Service = (*Repository)(nil)

	_ repoCommand = (*processMessageCmd)(nil)
	_ repoCommand = (*inspectCmd)(nil)
	_ repoCommand = (*removeCmd)(nil)
)

func NewRepository(emitter *myEvents.Emitter, connFactory connections.Factory) (r *Repository) {
	r = &Repository{
		Emitter:     emitter,
		ConnFactory: connFactory,
		commands:    make(chan repoCommand, 50),
		containers:  make(map[ID]*Container, 10),
	}
	emitter.OnNewReceiver = r.primeNewReceiver
	return r
}

func (r *Repository) Serve(ctx context.Context) (err error) {
	r.conn, err = r.ConnFactory.CreateConn()
	if err != nil {
		return
	}
	defer func() {
		_ = r.conn.Close()
		r.conn = nil
	}()

	for err == nil {
		select {
		case cmd := <-r.commands:
			err = cmd.Execute(r, ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return
}

func (r *Repository) Process(msg events.Message) {
	r.commands <- processMessageCmd{msg}
}

func (r *Repository) primeNewReceiver(receiver myEvents.Receiver) {
	ctx := context.Background()
	for _, c := range r.containers {
		event := myEvents.MakeEvent(myEvents.ContainerTarget, string(c.ID), myEvents.TargetUpdated, c.LastUpdateTime(), c)
		receiver.Receive(event, ctx)
	}
}

func (c processMessageCmd) Execute(r *Repository, ctx context.Context) error {
	when := time.Unix(0, c.msg.TimeNano)
	switch {
	case c.msg.Type == "container" && c.msg.Action == "destroy":
		r.commands <- removeCmd{ID(c.msg.ID), when}
	case c.msg.Type == "container":
		r.commands <- inspectCmd{ID(c.msg.ID), when}
	case c.msg.Type == "network" && c.msg.Action == "connect",
		c.msg.Type == "network" && c.msg.Action == "disconnect":
		r.commands <- inspectCmd{ID(c.msg.Actor.Attributes["containers"]), when}
	}
	return nil
}

func (c inspectCmd) Execute(r *Repository, ctx context.Context) error {
	ctn, found := r.containers[c.id]
	if found && ctn.IsRemoved() {
		return nil
	}
	data, err := r.conn.ContainerInspect(ctx, string(c.id))
	if err != nil {
		return err
	}
	if !found {
		ctn = NewContainer(data)
		r.containers[c.id] = ctn
		log.Printf("added container: %s", c.id)
	} else {
		ctn.Update(data)
		log.Printf("updated container: %s", c.id)
	}
	r.Emitter.Emit(myEvents.MakeEvent(myEvents.ContainerTarget, string(c.id), myEvents.TargetUpdated, c.when, ctn))
	return nil
}

func (c removeCmd) Execute(r *Repository, ctx context.Context) error {
	ctn, found := r.containers[c.id]
	if !found {
		return nil
	}
	ctn.RemovedAt = c.when
	delete(r.containers, c.id)
	log.Printf("removed container: %s", c.id)
	r.Emitter.Emit(myEvents.MakeEvent(myEvents.ContainerTarget, string(c.id), myEvents.TargetRemoved, c.when, nil))
	return nil
}
