package membership

import (
	"encoding/json"

	"github.com/sudame/chat/internal/events"
)

const MembershipCreatedEventType string = "membership.created"

type MembershipCreatedEvent struct {
	Id         string `json:"id"`
	UserId     string `json:"user_id"`
	ChatRoomId string `json:"chat_room_id"`
}

// ToEnvelope implements [events.Event].
func (e *MembershipCreatedEvent) ToEnvelope() (*events.EventEnvelope, error) {

	payload, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return &events.EventEnvelope{
		Type:    MembershipCreatedEventType,
		Payload: payload,
	}, nil
}

var _ events.Event = (*MembershipCreatedEvent)(nil)
