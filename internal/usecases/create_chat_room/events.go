package createchatroom

type ChatRoomCreatedEvent struct {
	ChatRoomID int64
	Name       string
}

func (e *ChatRoomCreatedEvent) EventType() string {
	return "ChatRoomCreated"
}
