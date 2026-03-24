package message

import (
	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/events"
)

type Message struct {
	ID           string         `json:"id"`
	ChatRoomID   string         `json:"chat_room_id"`
	AuthorUserID string         `json:"author_user_id"`
	Body         string         `json:"body"`
	Events       []events.Event `json:"-"`
}

func PostMessage(chatRoomID string, authorUserID string, body string) (*Message, error) {
	id := "message:" + ulid.Make().String()

	event := MessagePostedEvent{
		ID:           id,
		AuthorUserID: authorUserID,
		ChatRoomID:   chatRoomID,
		Body:         body,
	}

	return &Message{
		ID:           id,
		AuthorUserID: authorUserID,
		ChatRoomID:   chatRoomID,
		Body:         body,
		Events:       []events.Event{&event},
	}, nil
}

func ReconstructMessage(id string, chatRoomID string, authorUserID string, body string) *Message {
	return &Message{
		ID:           id,
		ChatRoomID:   chatRoomID,
		AuthorUserID: authorUserID,
		Body:         body,
		Events:       []events.Event{},
	}
}

func (m *Message) ClearEvents() {
	m.Events = []events.Event{}
}
