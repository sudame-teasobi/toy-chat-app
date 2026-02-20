package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/internal/domain/user"
	"github.com/sudame/chat/internal/ticdc"
)

type ReadModelConsumer struct {
	client *dynamodb.Client
}

func NewReadModelConsumer(client *dynamodb.Client) *ReadModelConsumer {
	return &ReadModelConsumer{
		client: client,
	}
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

type User struct {
	Id   string `dynamodbav:"id"`
	Name string `dynamodbav:"name"`
}

type Room struct {
	Id   string `dynamodbav:"id"`
	Name string `dynamodbav:"name"`
}

type Membership struct {
	Id     string `dynamodbav:"id"`
	RoomID string `dynamodbav:"room_id"`
	UserID string `dynamodbav:"user_id"`
}

func handleUserCreatedEvent(ctx context.Context, client *dynamodb.Client, event user.UserCreatedEvent) error {
	user := User{Id: event.UserID, Name: event.Name}
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to construct attribute: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{Item: item, TableName: new("Users")})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

func handleRoomCreatedEvent(ctx context.Context, client *dynamodb.Client, event room.ChatRoomCreatedEvent) error {
	room := Room{Id: event.ChatRoomID, Name: event.Name}
	item, err := attributevalue.MarshalMap(room)
	if err != nil {
		return fmt.Errorf("failed to construct attribute: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{Item: item, TableName: new("Rooms")})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

func handleMembershipCreatedEvent(ctx context.Context, client *dynamodb.Client, event membership.MembershipCreatedEvent) error {
	// FIXIT: membershipID をここで生成するのは意味不明なので要修正
	membershipID := "membership:" + ulid.Make().String()
	membership := Membership{Id: membershipID, RoomID: event.ChatRoomId, UserID: event.UserId}
	item, err := attributevalue.MarshalMap(membership)
	if err != nil {
		return fmt.Errorf("failed to construct attribute: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{Item: item, TableName: new("Memberships")})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

// TODO: kafka は at-least-once な保証スタイルなので、重複したイベントが飛んできたときの処理を検討すべき
// Consume はコンシュームする
func (c *ReadModelConsumer) Consume(ctx context.Context, event ticdc.Event) error {
	for _, data := range event.Data {
		var es string
		err := json.Unmarshal(data.Payload, &es)
		if err != nil {
			return fmt.Errorf("failed to construct payload to string: %w", err)
		}
		eb := []byte(es)
		switch data.Type {
		case room.ChatRoomCreatedEventType:
			devent, err := extractEventData[room.ChatRoomCreatedEvent](eb)
			if err != nil {
				return fmt.Errorf("failed to extract domain event: %w", err)
			}
			err = handleRoomCreatedEvent(ctx, c.client, devent)
			if err != nil {
				return fmt.Errorf("failed to handle room created event: %w", err)
			}
			slog.DebugContext(ctx, "room created", "domain_data", devent)
		case user.UserCreatedEventType:
			devent, err := extractEventData[user.UserCreatedEvent](eb)
			if err != nil {
				return fmt.Errorf("failed to extract domain event: %w", err)
			}
			err = handleUserCreatedEvent(ctx, c.client, devent)
			if err != nil {
				return fmt.Errorf("failed to handle user created event: %w", err)
			}
			slog.DebugContext(ctx, "user created", "domain_data", devent)
		case membership.MembershipCreatedEventType:
			devent, err := extractEventData[membership.MembershipCreatedEvent](eb)
			if err != nil {
				return fmt.Errorf("failed to extract domain event: %w", err)
			}
			err = handleMembershipCreatedEvent(ctx, c.client, devent)
			if err != nil {
				return fmt.Errorf("failed to handle membership created event: %w", err)
			}
			slog.DebugContext(ctx, "membership created", "domain_data", devent)
		default:
			return fmt.Errorf("failed to handle event, unknown event type: event_type = %s", data.Type)
		}

	}

	return nil
}
