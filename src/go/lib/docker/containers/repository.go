package containers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/adirelle/docker-graph/src/go/lib/docker/connections"
	myEvents "github.com/adirelle/docker-graph/src/go/lib/docker/events"
	"github.com/adirelle/docker-graph/src/go/lib/utils"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	log "github.com/inconshreveable/log15"
	"github.com/thejerf/suture/v4"
)

var (
	Log       = log.New()
	LoggerKey = struct{}{}
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
	logger := Log.New(log.Ctx{"id": msg.ID})
	msgLogger := logger.New(log.Ctx{
		"type":   msg.Type,
		"action": msg.Action,
	})
	subCtx := context.WithValue(ctx, LoggerKey, logger)
	if r.doHandleMessage(msg, subCtx) {
		msgLogger.Debug("handled message")
	} else {
		msgLogger.Debug("ignored message")
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
					// inspectCtx, cancel := context.WithTimeout(
					// 	context.WithValue(context.Background(), LoggerKey, Log.New("id", id)),
					// 	5*time.Second,
					// )
					// defer cancel()
					// r.inspectContainer(t, when, inspectCtx)
				},
				Delay: InspectTimeout,
			},
		}
		r.containers[id] = t
		ctx.Value(LoggerKey).(log.Logger).Info("added container")
	}
	//t.Trigger()
	r.inspectContainer(t, when, ctx)
}

func (r *Repository) inspectContainer(t *tracker, when time.Time, ctx context.Context) {
	Log.Debug("inspecting container", "id", t.ID)
	data, err := r.conn.ContainerInspect(ctx, string(t.ID))
	if err != nil {
		if !client.IsErrNotFound(err) {
			Log.Error("errror inspecting container", "id", t.ID, "error", err)
		}
		return
	}
	Log.Debug("updating container", "id", t.ID)
	t.Update(data, when)
	if t.IsRemoved() {
		r.removeContainer(t.ID, when, ctx)
	} else {
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
	Log.Debug("removed container")
	r.Emitter.Emit(myEvents.MakeContainerRemovedEvent(id, when))
}
