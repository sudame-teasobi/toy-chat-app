package user

import (
	"encoding/json"

	"github.com/sudame/chat/internal/events"
)

var _ events.Event = (*UserCreatedEvent)(nil)

// UserCreatedEvent raised when a useris created.
type UserCreatedEvent struct {
	UserID string
	Name   string
}

// ToEnvelope implements [events.Event].
func (u *UserCreatedEvent) ToEnvelope() (*events.EventEnvelope, error) {
	payload, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	return &events.EventEnvelope{
		Type:    "user.created",
		Payload: payload,
	}, nil
}
