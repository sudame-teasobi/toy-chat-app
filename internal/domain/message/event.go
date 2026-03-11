package message

import (
	"encoding/json"

	"github.com/sudame/chat/internal/events"
)

const MessagePostedEventType string = "messsage.posted"

type MessagePostedEvent struct {
	ID           string `json:"id"`
	AuthorUserID string `json:"author_user_id"`
	ChatRoomID   string `json:"chat_room_id"`
	Body         string `json:"body"`
}

// ToEnvelope implements [events.Event].
func (e *MessagePostedEvent) ToEnvelope() (*events.EventEnvelope, error) {

	payload, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return &events.EventEnvelope{
		Type:    MessagePostedEventType,
		Payload: payload,
	}, nil
}

var _ events.Event = (*MessagePostedEvent)(nil)
