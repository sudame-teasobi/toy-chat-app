package room

import (
	"encoding/json"

	"github.com/sudame/chat/internal/events"
)

var (
	_ events.Event = (*ChatRoomCreatedEvent)(nil)
)

const ChatRoomCreatedEventType string = "chatroom.created"

// ChatRoomCreatedEvent is raised when a chat room is created.
type ChatRoomCreatedEvent struct {
	ChatRoomID    string `json:"chat_room_id"`
	Name          string `json:"name"`
	CreatorUserID string `json:"creator_user_id"`
}

func (e *ChatRoomCreatedEvent) ToEnvelope() (*events.EventEnvelope, error) {
	payload, err := json.Marshal(e)

	if err != nil {
		return nil, err
	}

	return &events.EventEnvelope{
		Type:    ChatRoomCreatedEventType,
		Payload: payload,
	}, nil
}
