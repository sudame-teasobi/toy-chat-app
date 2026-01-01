package events

// Event is the interface for all domain events.
type Event interface {
	EventType() string
}
