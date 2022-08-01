package containers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/adirelle/docker-graph/src/go/lib/docker/connections"
	myEvents "github.com/adirelle/docker-graph/src/go/lib/docker/events"
	"github.com/adirelle/docker-graph/src/go/lib/utils"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/thejerf/suture/v4"
)

type (
	Repository struct {
		Emitter     *myEvents.Emitter
		ConnFactory connections.Factory

		conn       connections.Connection
		messages   chan events.Message
		containers map[ID]*tracker
	}

	tracker struct {
		Container
		utils.Debouncer
	}
)

var (
	_ suture.Service = (*Repository)(nil)
	_ fmt.GoStringer = (*Repository)(nil)

	InspectTimeout = 200 * time.Millisecond
)

func NewRepository(emitter *myEvents.Emitter, connFactory connections.Factory) (r *Repository) {
	r = &Repository{
		Emitter:     emitter,
		ConnFactory: connFactory,
		messages:    make(chan events.Message, 50),
		containers:  make(map[ID]*tracker, 10),
	}
	emitter.OnNewReceiver = r.primeNewReceiver
	return r
}

func (r *Repository) GoString() string {
	return fmt.Sprintf("containers.Repository(%d, %d/%d)", len(r.containers), len(r.messages), cap(r.messages))
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
		case msg := <-r.messages:
			err = r.handleMessage(msg, ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return
}

func (r *Repository) Process(msg events.Message) {
	r.messages <- msg
}

func (r *Repository) primeNewReceiver(receiver myEvents.Receiver) {
	ctx := context.Background()
	for _, t := range r.containers {
		if !t.UpdatedAt.IsZero() {
			event := myEvents.MakeContainerUpdatedEvent(t.ID, t.UpdatedAt, t.Container)
			receiver.Receive(event, ctx)
		}
	}
}

func (r *Repository) handleMessage(msg events.Message, ctx context.Context) error {
	if r.doHandleMessage(msg, ctx) {
		log.Printf("handled message %s:%s(%s)", msg.Type, msg.Action, msg.ID)
	} else {
		log.Printf("ignored %s:%s message", msg.Type, msg.Action)
	}
	return nil
}

func (r *Repository) doHandleMessage(msg events.Message, ctx context.Context) bool {
	when := time.Unix(0, msg.TimeNano)
	switch msg.Type {
	case "container":
		if msg.Action == "destroy" {
			r.removeContainer(ID(msg.ID), when, ctx)
			return true
		}
		if msg.Action == "attach" || msg.Action == "detach" || strings.HasPrefix(msg.Action, "exec_") {
			return false
		}
		r.updateContainer(ID(msg.ID), when, ctx)
		return true
	case "network":
		if msg.Action == "connect" || msg.Action == "disconnect" {
			r.updateContainer(ID(msg.Actor.Attributes["containers"]), when, ctx)
			return true
		}
	}
	return false
}

func (r *Repository) updateContainer(id ID, when time.Time, ctx context.Context) {
	if id == "" {
		return
	}
	t, found := r.containers[id]
	if !found {
		t = &tracker{
			Container: Container{ID: id},
			Debouncer: utils.Debouncer{
				Func: func() {
					inspectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					r.inspectContainer(t, when, inspectCtx)
				},
				Delay: InspectTimeout,
			},
		}
		r.containers[id] = t
		log.Printf("added container: %s", id)
	}
	t.Trigger()
}

func (r *Repository) inspectContainer(t *tracker, when time.Time, ctx context.Context) {
	log.Printf("inspecting container: %s", t.ID)
	data, err := r.conn.ContainerInspect(ctx, string(t.ID))
	if err != nil {
		if !client.IsErrNotFound(err) {
			log.Println(err)
		}
		return
	}
	log.Printf("updating container: %s, %#v", t.ID, data)
	changed := t.Update(data, when)
	if t.IsRemoved() {
		r.removeContainer(t.ID, when, ctx)
	} else if changed {
		r.Emitter.Emit(myEvents.MakeContainerUpdatedEvent(t.ID, when, t.Container))
	}
}

func (r *Repository) removeContainer(id ID, when time.Time, ctx context.Context) {
	t, found := r.containers[id]
	if !found {
		return
	}
	t.Stop()
	delete(r.containers, id)
	t.Container.RemovedAt = when
	log.Printf("removed container: %s", id)
	r.Emitter.Emit(myEvents.MakeContainerRemovedEvent(id, when))
}