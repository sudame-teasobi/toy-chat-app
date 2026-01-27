package chatroom

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
	ChatRoomID    string `json:"chatRoomId"`
	Name          string `json:"name"`
	CreatorUserID string `json:"creatorUserId"`
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
