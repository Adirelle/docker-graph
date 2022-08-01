package listeners

import (
	"context"
	"fmt"
	"time"

	"github.com/adirelle/docker-graph/src/go/lib/docker/connections"
	"github.com/adirelle/docker-graph/src/go/lib/docker/containers"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/thejerf/suture/v4"

	log "github.com/inconshreveable/log15"
)

type (
	Listener struct {
		connFactory     connections.Factory
		repository      *containers.Repository
		lastMessageTime time.Time
	}
)

var (
	_ suture.Service = (*Listener)(nil)
	_ fmt.GoStringer = (*Listener)(nil)

	Log = log.New()
)

func NewListener(connFactory connections.Factory, repository *containers.Repository) *Listener {
	return &Listener{
		connFactory: connFactory,
		repository:  repository,
	}
}

func (m *Listener) GoString() string {
	return "Listener"
}

func (m *Listener) Serve(ctx context.Context) error {
	conn, err := m.connFactory.CreateConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Prime with existing containers
	if m.lastMessageTime.IsZero() {
		if err := m.prime(ctx, conn); err != nil {
			return err
		}
	}

	// Listen for events
	eventC, errC := conn.Events(ctx, types.EventsOptions{Since: m.lastMessageTime.Format(time.RFC3339)})
	for {
		select {
		case msg := <-eventC:
			Log.Debug("received message", "type", msg.Type, "action", msg.Action, "actor_id", msg.Actor.ID)
			m.lastMessageTime = time.Unix(0, msg.TimeNano)
			m.repository.Process(msg)
		case err = <-errC:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *Listener) prime(ctx context.Context, conn connections.Connection) error {
	containers, err := conn.ContainerList(ctx, types.ContainerListOptions{All: true, Since: "1"})
	if err != nil {
		return err
	}
	for i, ctn := range containers {
		created := time.Unix(ctn.Created, 0)
		if i == 0 || created.After(m.lastMessageTime) {
			m.lastMessageTime = created
		}
		m.repository.Process(events.Message{Type: "container", Action: "create", ID: ctn.ID, TimeNano: created.UnixNano()})
	}
	return nil
}
