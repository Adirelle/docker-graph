package docker

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/adirelle/docker-graph/web"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/thejerf/suture/v4"
)

type (
	StreamFactory struct {
		*suture.Supervisor
		connFactory ConnectionFactory
	}

	Stream struct {
		conn        Connection
		events      chan<- web.Event
		lastEventID EventID
	}

	Event struct {
		EventID       EventID
		TargetType    string
		TargetID      string
		Action        string
		Message       *events.Message
		ContainerInfo *types.ContainerJSON
	}

	EventID int64
)

var (
	_ suture.Service  = (*StreamFactory)(nil)
	_ web.EventSource = (*StreamFactory)(nil)

	_ suture.Service = (*Stream)(nil)

	_ web.Event = (*Event)(nil)

	_ web.EventID  = (*EventID)(nil)
	_ fmt.Stringer = (*EventID)(nil)

	eventFilter = map[string]map[string]bool{
		"container": {
			"create":  true,
			"start":   true,
			"kill":    true,
			"die":     true,
			"stop":    true,
			"destroy": true,
		},
		"network": {
			"connect":     true,
			"discconnect": true,
		},
	}
)

func NewStreamFactory(connFactory ConnectionFactory) *StreamFactory {
	return &StreamFactory{
		Supervisor:  suture.NewSimple("monitors"),
		connFactory: connFactory,
	}
}

func (f *StreamFactory) Listen(from web.EventID) (events <-chan web.Event, stop func(), err error) {
	eventChan := make(chan web.Event, 10)
	events = eventChan
	stream := &Stream{events: eventChan}
	if from != nil {
		if err = stream.lastEventID.FromString(from.String()); err != nil {
			return
		}
	}

	if stream.conn, err = f.connFactory.CreateConn(); err != nil {
		return
	}
	svcToken := f.Supervisor.Add(stream)
	stop = func() { _ = f.Supervisor.Remove(svcToken) }
	return
}

func (m *Stream) Serve(ctx context.Context) (err error) {
	opts := types.EventsOptions{}
	if !m.lastEventID.IsZero() {
		opts.Since = m.lastEventID.Time().Format(time.RFC3339Nano)
	}

	var event *Event
	messages, errC := m.conn.Events(ctx, opts)
	for err == nil {
		select {
		case message := <-messages:
			if event, err = m.processMessage(&message); event == nil && err == nil {
				log.Printf("send: %#v", event)
				select {
				case m.events <- event:
				case <-ctx.Done():
					err = ctx.Err()
				}
			}
		case err = <-errC:
		case <-ctx.Done():
			err = ctx.Err()
		}
	}
	return
}

func (m *Stream) processMessage(msg *events.Message) (event *Event, err error) {
	if actionFilter, found := eventFilter[msg.Type]; !found {
		return
	} else if accepted := actionFilter[msg.Action]; !accepted {
		return
	}

	event = &Event{
		TargetType: msg.Type,
		TargetID:   msg.ID,
		Action:     msg.Action,
		Message:    msg,
	}
	if err = event.EventID.FromTimestampNano(msg.TimeNano); err != nil {
		return
	}

	return
}

func (e *Event) Type() string {
	return fmt.Sprintf("%s/%s", e.TargetType, e.Action)
}

func (e *Event) ID() web.EventID {
	return e.EventID
}

func (e *Event) Payload() interface{} {
	return e
}

func (i EventID) IsZero() bool {
	return int64(i) == 0
}

func (i EventID) IsGreaterThan(other EventID) bool {
	return int64(i) > int64(other)
}

func (i EventID) IsLowerThan(other EventID) bool {
	return int64(i) < int64(other)
}

func (i *EventID) FromString(input string) error {
	if value, err := strconv.ParseInt(input, 10, 64); err == nil {
		return i.FromTimestampNano(value)
	} else {
		return err
	}
}

func (i *EventID) FromTimestamp(timestamp int64) error {
	return i.FromTimestampNano(timestamp * 1e9)
}

func (i *EventID) FromTimestampNano(timestampNano int64) error {
	*i = EventID(timestampNano)
	return nil
}

func (i EventID) String() string {
	return strconv.FormatInt(i.TimestampNano(), 10)
}

func (i EventID) TimestampNano() int64 {
	return int64(i)
}

func (i EventID) Time() time.Time {
	return time.Unix(0, i.TimestampNano())
}
