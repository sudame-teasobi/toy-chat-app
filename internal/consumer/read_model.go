package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/internal/domain/user"
	"github.com/sudame/chat/internal/service"
	"github.com/sudame/chat/internal/ticdc"
)

type ReadModelConsumer struct{}

func NewReadModelConsumer(service *service.CreateMembershipService) *ReadModelConsumer {
	return &ReadModelConsumer{}
}

func extractEventData[EventType any](data json.RawMessage) (EventType, error) {
	var zero EventType
	var devent EventType
	err := json.Unmarshal(data, &devent)
	if err != nil {
		return zero, fmt.Errorf("failed to extract event: %w", err)
	}
	return devent, nil
}

// TODO: kafka は at-least-once な保証スタイルなので、重複したイベントが飛んできたときの処理を検討すべき
// Consume はコンシュームする
func (c *ReadModelConsumer) Consume(ctx context.Context, event ticdc.Event) error {
	for _, data := range event.Data {

		switch data.Type {
		case room.ChatRoomCreatedEventType:
			devent, err := extractEventData[room.ChatRoomCreatedEvent](data.Payload)
			if err != nil {
				slog.ErrorContext(ctx, "failed to extract domain event", "error", err)
			}
			slog.DebugContext(ctx, "room created", "domain_data", devent)
		case user.UserCreatedEventType:
			devent, err := extractEventData[user.UserCreatedEvent](data.Payload)
			if err != nil {
				slog.ErrorContext(ctx, "failed to extract domain event", "error", err)
			}
			slog.DebugContext(ctx, "user created", "domain_data", devent)
		case membership.MembershipCreatedEventType:
			devent, err := extractEventData[membership.MembershipCreatedEvent](data.Payload)
			if err != nil {
				slog.ErrorContext(ctx, "failed to extract domain event", "error", err)
			}
			slog.DebugContext(ctx, "membership created", "domain_data", devent)
		default:
			slog.ErrorContext(ctx, "failed to handle event, unknown event type", "event_type", data.Type)
		}

	}

	return nil
}
