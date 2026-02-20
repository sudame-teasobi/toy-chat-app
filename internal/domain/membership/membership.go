package membership

import (
	"encoding/json"

	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/events"
)

const MembershipCreatedEventType string = "membership.created"

type MembershipCreatedEvent struct {
	Id         string `json:"id"`
	UserId     string `json:"user_id"`
	ChatRoomId string `json:"chat_room_id"`
}

// ToEnvelope implements [events.Event].
func (e *MembershipCreatedEvent) ToEnvelope() (*events.EventEnvelope, error) {

	payload, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return &events.EventEnvelope{
		Type:    MembershipCreatedEventType,
		Payload: payload,
	}, nil
}

var _ events.Event = (*MembershipCreatedEvent)(nil)

type Membership struct {
	Id         string         `json:"id"`
	ChatRoomId string         `json:"chatRoomId"`
	UserId     string         `json:"userId"`
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
