package user

import "github.com/sudame/chat/internal/events"

var _ events.Event = (*UserCreatedEvent)(nil)

// UserCreatedEvent raised when a useris created.
type UserCreatedEvent struct {
	UserID int64
	Name   string
}

func (e *UserCreatedEvent) EventType() string {
	return "UserCreated"
}
