package events

import "encoding/json"

type EventEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Event is the interface for all domain events.
type Event interface {
	EventType() string
}

type EventHandler interface {
	Handle(event any) error
}
