package events

type Event interface {
	EventType() string
}

type MemberAddedEvent struct {
	ChatRoomID int64
	UserID     int64
}

func (e *MemberAddedEvent) EventType() string {
	return "MemberAdded"
}
