package docker

import (
	"context"
	"time"

	"github.com/adirelle/docker-graph/pkg/lib/docker/connections"
	"github.com/adirelle/docker-graph/pkg/lib/docker/containers"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/thejerf/suture/v4"
)

type (
	MessageConsumer struct {
		connFactory     connections.Factory
		repository      *containers.Repository
		lastMessageTime time.Time
	}
)

var _ suture.Service = (*MessageConsumer)(nil)

func NewMessageConsumer(connFactory connections.Factory, repository *containers.Repository) *MessageConsumer {
	return &MessageConsumer{
		connFactory: connFactory,
		repository:  repository,
	}
}

func (m *MessageConsumer) Serve(ctx context.Context) error {
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
			// log.Printf("received %s:%s message (%s)", msg.Type, msg.Action, msg.Actor.ID)
			m.lastMessageTime = time.Unix(0, msg.TimeNano)
			m.repository.Process(msg)
		case err = <-errC:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *MessageConsumer) prime(ctx context.Context, conn connections.Connection) error {
	containers, err := conn.ContainerList(ctx, types.ContainerListOptions{All: true, Since: "1"})
	if err != nil {
		return err
	}
	for i, ctn := range containers {
		created := time.Unix(ctn.Created, 0)
		if i == 0 || created.After(m.lastMessageTime) {
			m.lastMessageTime = created
		}
		m.repository.Process(events.Message{Type: "container", Action: "created", ID: ctn.ID, TimeNano: created.UnixNano()})
	}
	return nil
}
