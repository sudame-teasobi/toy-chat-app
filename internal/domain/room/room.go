package room

import (
	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/events"
)

// Room is the aggregate root for chat room domain.
type Room struct {
	id     string
	name   string
	events []events.Event
}

// NewRoom creates a new ChatRoom aggregate.
// It validates the name and records ChatRoomCreatedEvent.
// Member addition is handled asynchronously by the event consumer.
func NewRoom(name string, creatorUserID string) (*Room, error) {
	id := "chat-room:" + ulid.Make().String()

	if name == "" {
		return nil, ErrEmptyName
	}

	chatroom := &Room{
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

// ReconstructRoom reconstructs a ChatRoom from persistence.
func ReconstructRoom(id string, name string) *Room {
	return &Room{
		id:     id,
		name:   name,
		events: make([]events.Event, 0),
	}
}

func (cr *Room) ID() string   { return cr.id }
func (cr *Room) Name() string { return cr.name }

// Events returns all recorded domain events.
func (cr *Room) Events() []events.Event {
	return cr.events
}

// ClearEvents clears recorded events after persistence.
func (cr *Room) ClearEvents() {
	cr.events = make([]events.Event, 0)
}
