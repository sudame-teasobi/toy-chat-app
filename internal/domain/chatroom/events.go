package chatroom

import (
	"github.com/sudame/chat/internal/events"
)

var (
	_ events.Event = (*ChatRoomCreatedEvent)(nil)
)

// ChatRoomCreatedEvent is raised when a chat room is created.
type ChatRoomCreatedEvent struct {
	ChatRoomID    string `json:"chatRoomId"`
	Name          string `json:"name"`
	CreatorUserID string `json:"creatorUserId"`
}

func (e *ChatRoomCreatedEvent) EventType() string {
	return "chatroom.created"
}
