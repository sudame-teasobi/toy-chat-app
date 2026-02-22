package membership

import (
	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/events"
)

type Membership struct {
	Id         string         `json:"id"`
	ChatRoomId string         `json:"chat_room_id"`
	UserId     string         `json:"user_id"`
	Events     []events.Event `json:"-"`
}

func CreateMembership(chatRoomId string, userId string) (*Membership, error) {
	id := "membership-" + ulid.Make().String()

	event := MembershipCreatedEvent{
		Id:         id,
		UserId:     userId,
		ChatRoomId: chatRoomId,
	}

	return &Membership{
		Id:         id,
		ChatRoomId: chatRoomId,
		UserId:     userId,
		Events:     []events.Event{&event},
	}, nil
}
