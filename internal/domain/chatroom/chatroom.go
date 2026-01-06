package chatroom

import (
	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/events"
)

// ChatRoom is the aggregate root for chat room domain.
type ChatRoom struct {
	id      string
	name    string
	members []Member
	events  []events.Event
}

// NewChatRoom creates a new ChatRoom aggregate.
// It validates the name and records ChatRoomCreatedEvent.
// Member addition is handled asynchronously by the event consumer.
func NewChatRoom(name string, creatorUserID string) (*ChatRoom, error) {
	id := "chat-room:" + ulid.Make().String()

	if name == "" {
		return nil, ErrEmptyName
	}

	cr := &ChatRoom{
		id:      id,
		name:    name,
		members: []Member{},
		events:  make([]events.Event, 0),
	}

	cr.events = append(cr.events, &ChatRoomCreatedEvent{
		ChatRoomID:    id,
		Name:          name,
		CreatorUserID: creatorUserID,
	})

	return cr, nil
}

// ReconstructChatRoom reconstructs a ChatRoom from persistence.
func ReconstructChatRoom(id string, name string, members []Member) *ChatRoom {
	return &ChatRoom{
		id:      id,
		name:    name,
		members: members,
		events:  make([]events.Event, 0),
	}
}

// AddMember adds a new member to the chat room.
func (cr *ChatRoom) AddMember(userID string) error {
	if cr.IsMember(userID) {
		return ErrAlreadyMember
	}

	cr.members = append(cr.members, NewMember(userID))
	cr.events = append(cr.events, &MemberAddedEvent{
		ChatRoomID: cr.id,
		UserID:     userID,
	})

	return nil
}

// IsMember checks if the user is already a member.
func (cr *ChatRoom) IsMember(userID string) bool {
	for _, m := range cr.members {
		if m.UserID() == userID {
			return true
		}
	}
	return false
}

func (cr *ChatRoom) ID() string   { return cr.id }
func (cr *ChatRoom) Name() string { return cr.name }
func (cr *ChatRoom) Members() []Member {
	result := make([]Member, len(cr.members))
	copy(result, cr.members)
	return result
}

// Events returns all recorded domain events.
func (cr *ChatRoom) Events() []events.Event {
	return cr.events
}

// ClearEvents clears recorded events after persistence.
func (cr *ChatRoom) ClearEvents() {
	cr.events = make([]events.Event, 0)
}
