package membership

import (
	"context"
	"encoding/json"

	"github.com/sudame/chat/internal/domain/chatroom"
	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/internal/ticdc"
)

func Listen(ctx context.Context, value []byte) {

	var ticdcevent ticdc.Event
	err := json.Unmarshal(value, &ticdcevent)
	if err != nil {
		// TODO: エラーハンドリング
		panic(nil)
	}

	switch ticdcevent.Data.Type {
	case "chatroom.created":
		var chatRoomCreatedEvent chatroom.ChatRoomCreatedEvent
		err := json.Unmarshal(ticdcevent.Data.Payload, &chatRoomCreatedEvent)
		if err != nil {
			// TODO: error handling
			panic(nil)
		}
		HandleChatRoomCreatedEvent(ctx, &chatRoomCreatedEvent)
	}
}

var membershipRepository membership.Repository

// TODO: 他のファイルに移す
func HandleChatRoomCreatedEvent(ctx context.Context, event *chatroom.ChatRoomCreatedEvent) {

	m, err := membership.NewMembership(event.ChatRoomID, event.CreatorUserID)
	if err != nil {
		// TODO: エラーハンドリング
		panic(nil)
	}

	err = membershipRepository.Save(ctx, m)
	if err != nil {
		// TODO: error handling
		panic(nil)
	}
}
