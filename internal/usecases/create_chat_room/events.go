package createchatroom

type Event interface {
	EventType() string
}

type ChatRoomCreatedEvent struct {
	ChatRoomID int64
	Name       string
}

func (e *ChatRoomCreatedEvent) EventType() string {
	return "ChatRoomCreated"
}

type MemberAddedEvent struct {
	ChatRoomID int64
	UserID     int64
}

func (e *MemberAddedEvent) EventType() string {
	return "MemberAdded"
}
