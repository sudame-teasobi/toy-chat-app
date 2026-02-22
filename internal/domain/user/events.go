package user

import (
	"encoding/json"

	"github.com/sudame/chat/internal/events"
)

var _ events.Event = (*UserCreatedEvent)(nil)

const UserCreatedEventType string = "user.created"

// UserCreatedEvent raised when a useris created.
type UserCreatedEvent struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

// ToEnvelope implements [events.Event].
func (u *UserCreatedEvent) ToEnvelope() (*events.EventEnvelope, error) {
	payload, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	return &events.EventEnvelope{
		Type:    UserCreatedEventType,
		Payload: payload,
	}, nil
}
