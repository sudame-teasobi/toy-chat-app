package chatroom

import (
	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/events"
)

// ChatRoom is the aggregate root for chat room domain.
type ChatRoom struct {
	id     string
	name   string
	events []events.Event
}

// NewChatRoom creates a new ChatRoom aggregate.
// It validates the name and records ChatRoomCreatedEvent.
// Member addition is handled asynchronously by the event consumer.
func NewChatRoom(name string, creatorUserID string) (*ChatRoom, error) {
	id := "chat-room:" + ulid.Make().String()

	if name == "" {
		return nil, ErrEmptyName
	}

	chatroom := &ChatRoom{
		id:     id,
		name:   name,
		events: make([]events.Event, 0),
	}

	chatroom.events = append(chatroom.events, &ChatRoomCreatedEvent{
		ChatRoomID:    id,
		Name:          name,
		CreatorUserID: creatorUserID,
	})

	return chatroom, nil
}

// ReconstructChatRoom reconstructs a ChatRoom from persistence.
func ReconstructChatRoom(id string, name string) *ChatRoom {
	return &ChatRoom{
		id:     id,
		name:   name,
		events: make([]events.Event, 0),
	}
}

func (cr *ChatRoom) ID() string   { return cr.id }
func (cr *ChatRoom) Name() string { return cr.name }

// Events returns all recorded domain events.
func (cr *ChatRoom) Events() []events.Event {
	return cr.events
}

// ClearEvents clears recorded events after persistence.
func (cr *ChatRoom) ClearEvents() {
	cr.events = make([]events.Event, 0)
}
