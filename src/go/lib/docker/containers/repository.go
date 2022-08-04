package containers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/adirelle/docker-graph/src/go/lib/api"
	"github.com/adirelle/docker-graph/src/go/lib/docker/connections"
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
		ConnFactory connections.Factory

		conn       connections.Connection
		dispatcher Dispatcher
		messages   chan events.Message
		containers map[ID]*Container
	}

	Dispatcher interface {
		Dispatch(value api.Event, ctx context.Context) error
		OnNewSubscriber(hook func(chan<- api.Event))
	}
)

var (
	_ suture.Service = (*Repository)(nil)
	_ fmt.GoStringer = (*Repository)(nil)

	InspectTimeout = 200 * time.Millisecond
)

func NewRepository(dispatcher Dispatcher, connFactory connections.Factory) (r *Repository) {
	r = &Repository{
		dispatcher:  dispatcher,
		ConnFactory: connFactory,
		messages:    make(chan events.Message, 50),
		containers:  make(map[ID]*Container, 10),
	}
	dispatcher.OnNewSubscriber(r.primeNewSubscriber)
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

func (r *Repository) primeNewSubscriber(c chan<- api.Event) {
	Log.Debug("new subscriber", "c", c, "#ctn", len(r.containers))
	for _, ctn := range r.containers {
		Log.Debug("sending container", "ctn", ctn)
		c <- &ContainerUpdated{ctn.LastUpdateTime(), ctn}
	}
}

func (r *Repository) handleMessage(msg events.Message, ctx context.Context) error {
	logger := Log.New(log.Ctx{"id": msg.ID})
	ctx = context.WithValue(ctx, LoggerKey, logger)
	when := time.Unix(0, msg.TimeNano)
	switch msg.Type {
	case "container":
		if msg.Action == "destroy" {
			r.removeContainer(ID(msg.ID), when, ctx)
		} else if msg.Action == "attach" || msg.Action == "detach" || strings.HasPrefix(msg.Action, "exec_") {
			return nil
		} else {
			r.updateContainer(ID(msg.ID), when, ctx)
		}
	case "network":
		if msg.Action == "connect" || msg.Action == "disconnect" {
			r.updateContainer(ID(msg.Actor.Attributes["containers"]), when, ctx)
		}
	}
	return nil
}

func (r *Repository) updateContainer(id ID, when time.Time, ctx context.Context) {
	if id == "" {
		return
	}

	logger := ctx.Value(LoggerKey).(log.Logger)
	ctn, found := r.containers[id]
	if !found {
		ctn = &Container{ID: id, CreatedAt: when}
		r.containers[id] = ctn
		logger.Debug("added container")
	} else {
		logger.Debug("updating container")
	}

	data, err := r.conn.ContainerInspect(ctx, string(id))
	if err != nil {
		if !client.IsErrNotFound(err) {
			logger.Error("errror inspecting container", "error", err)
		}
		return
	}

	ctn.UpdateFrom(data)
	if ctn.Status.IsRemoved() {
		r.removeContainer(id, when, ctx)
	} else {
		ctn.UpdatedAt = when
		r.dispatcher.Dispatch(&ContainerUpdated{when, ctn}, ctx)
	}
}

func (r *Repository) removeContainer(id ID, when time.Time, ctx context.Context) {
	_, found := r.containers[id]
	if !found {
		return
	}
	delete(r.containers, id)
	logger := ctx.Value(LoggerKey).(log.Logger)
	logger.Debug("removed container")
	r.dispatcher.Dispatch(&ContainerRemoved{when, string(id)}, ctx)
}
