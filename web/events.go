package web

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

type (
	Streamer struct {
		events <-chan Event
		stop   func()
	}

	EventID interface {
		fmt.Stringer
	}

	Event interface {
		Type() string
		ID() EventID
		Payload() interface{}
	}

	EventSource interface {
		Listen(from EventID) (events <-chan Event, stop func(), err error)
	}

	eventId string
)

var _ EventID = (*eventId)(nil)

func (s *Streamer) start(output *bufio.Writer) {
	encoder := json.NewEncoder(output)

	defer s.stop()
	for event := range s.events {
		if err := s.writeEvent(output, encoder, event); err == io.EOF {
			return
		} else if err != nil {
			fmt.Printf("Error while writing event: %s", err)
		}
	}
}

func (s *Streamer) writeEvent(output *bufio.Writer, encoder *json.Encoder, event Event) (err error) {
	if _, err = fmt.Fprintf(output, "id:%s\nevent:%s\ndata:", event.ID(), event.Type()); err != nil {
		return
	}
	if err = encoder.Encode(event.Payload()); err == io.EOF {
		return
	}
	if _, err = output.WriteString("\n"); err != nil {
		return
	}
	return output.Flush()
}

func parseEventID(id string) eventId {
	return eventId(id)
}

func (s eventId) String() string {
	return string(s)
}
