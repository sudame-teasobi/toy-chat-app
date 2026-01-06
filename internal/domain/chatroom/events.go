package chatroom

import (
	"github.com/sudame/chat/internal/events"
)

var (
	_ events.Event = (*ChatRoomCreatedEvent)(nil)
	_ events.Event = (*MemberAddedEvent)(nil)
)

// ChatRoomCreatedEvent is raised when a chat room is created.
type ChatRoomCreatedEvent struct {
	ChatRoomID    string
	Name          string
	CreatorUserID string
}

func (e *ChatRoomCreatedEvent) EventType() string {
	return "ChatRoomCreated"
}

// MemberAddedEvent is raised when a member is added to a chat room.
type MemberAddedEvent struct {
	ChatRoomID string
	UserID     string
}

func (e *MemberAddedEvent) EventType() string {
	return "MemberAdded"
}
