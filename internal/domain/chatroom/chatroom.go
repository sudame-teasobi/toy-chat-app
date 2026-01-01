package chatroom

import "github.com/sudame/chat/internal/events"

// ChatRoom is the aggregate root for chat room domain.
type ChatRoom struct {
	id      int64
	name    string
	members []Member
	events  []events.Event
}

// NewChatRoom creates a new ChatRoom aggregate.
// It validates the name, adds the creator as the first member,
// and records ChatRoomCreatedEvent and MemberAddedEvent.
func NewChatRoom(id int64, name string, creatorUserID int64) (*ChatRoom, error) {
	if name == "" {
		return nil, ErrEmptyName
	}

	cr := &ChatRoom{
		id:      id,
		name:    name,
		members: []Member{NewMember(creatorUserID)},
		events:  make([]events.Event, 0),
	}

	cr.events = append(cr.events, &ChatRoomCreatedEvent{
		ChatRoomID: id,
		Name:       name,
	})
	cr.events = append(cr.events, &MemberAddedEvent{
		ChatRoomID: id,
		UserID:     creatorUserID,
	})

	return cr, nil
}

// ReconstructChatRoom reconstructs a ChatRoom from persistence.
func ReconstructChatRoom(id int64, name string, members []Member) *ChatRoom {
	return &ChatRoom{
		id:      id,
		name:    name,
		members: members,
		events:  make([]events.Event, 0),
	}
}

// AddMember adds a new member to the chat room.
func (cr *ChatRoom) AddMember(userID int64) error {
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
func (cr *ChatRoom) IsMember(userID int64) bool {
	for _, m := range cr.members {
		if m.UserID() == userID {
			return true
		}
	}
	return false
}

func (cr *ChatRoom) ID() int64        { return cr.id }
func (cr *ChatRoom) Name() string     { return cr.name }
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
