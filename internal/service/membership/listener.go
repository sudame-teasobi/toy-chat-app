package membership

import (
	"context"
	"encoding/json"

	"github.com/sudame/chat/internal/domain/chatroom"
	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/internal/ticdc"
)

func Listen(ctx context.Context, value []byte) error {

	var ticdcevent ticdc.Event
	err := json.Unmarshal(value, &ticdcevent)
	if err != nil {
		return err
	}

	switch ticdcevent.Data.Type {
	case membership.MembershipCreatedEventType:
		var chatRoomCreatedEvent chatroom.ChatRoomCreatedEvent
		err := json.Unmarshal(ticdcevent.Data.Payload, &chatRoomCreatedEvent)
		if err != nil {
			return err
		}

		err = HandleChatRoomCreatedEvent(ctx, &chatRoomCreatedEvent)
		if err != nil {
			return err
		}
	}

	return nil
}

var membershipRepository membership.Repository

// TODO: 他のファイルに移す
func HandleChatRoomCreatedEvent(ctx context.Context, event *chatroom.ChatRoomCreatedEvent) error {

	m, err := membership.CreateMembership(event.ChatRoomID, event.CreatorUserID)
	if err != nil {
		return err
	}

	err = membershipRepository.Save(ctx, m)
	if err != nil {
		return err
	}

	return nil
}
