package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/internal/service"
	"github.com/sudame/chat/internal/ticdc"
)

type MembershipConsumer struct {
	service *service.CreateMembershipService
}

func NewMembershipConsumer(service *service.CreateMembershipService) *MembershipConsumer {
	return &MembershipConsumer{
		service: service,
	}
}

// TODO: kafka は at-least-once な保証スタイルなので、重複したイベントが飛んできたときの処理を検討すべき
func (c *MembershipConsumer) Consume(ctx context.Context, event ticdc.Event) error {
	if event.Data.Type != room.ChatRoomCreatedEventType {
		return nil
	}

	var e room.ChatRoomCreatedEvent
	err := json.Unmarshal(event.Data.Payload, &e)
	if err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	err = c.service.Exec(ctx, e.CreatorUserID, e.ChatRoomID)
	if err != nil {
		return fmt.Errorf("failed to exec create-membership-service: %w", err)
	}

	return nil
}
