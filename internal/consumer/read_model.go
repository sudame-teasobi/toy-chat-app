package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/internal/domain/message"
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

type UserProfile struct {
	PK     string `dynamodbav:"PK"` // USER#<id>
	SK     string `dynamodbav:"SK"` // PROFILE
	UserID string `dynamodbav:"user_id"`
	Name   string `dynamodbav:"name"`
}

type RoomMetadata struct {
	PK     string `dynamodbav:"PK"` // ROOM#<id>
	SK     string `dynamodbav:"SK"` // METADATA
	RoomID string `dynamodbav:"room_id"`
	Name   string `dynamodbav:"name"`
}

type JoinedRoom struct {
	PK           string `dynamodbav:"PK"` // USER#<id>
	SK           string `dynamodbav:"SK"` // ROOM#<id>
	MembershipID string `dynamodbav:"membership_id"`
	RoomID       string `dynamodbav:"room_id"`
	UserID       string `dynamodbav:"user_id"`
}

type Membership struct {
	PK           string `dynamodbav:"PK"` // ROOM#<id>
	SK           string `dynamodbav:"SK"` // USER#<id>
	MembershipID string `dynamodbav:"membership_id"`
	RoomID       string `dynamodbav:"room_id"`
	UserID       string `dynamodbav:"user_id"`
}

type Message struct {
	PK        string `dynamodbav:"PK"` // ROOM#<id>
	SK        string `dynamodbav:"SK"` // MESSAGE#<id>
	MessageID string `dynamodbav:"message_id"`
	Body      string `dynamodbav:"body"`
	RoomID    string `dynamodbav:"room_id"`
	UserID    string `dynamodbav:"user_id"`
}

var tableName = "ToyChatApp"

func handleUserCreatedEvent(ctx context.Context, client *dynamodb.Client, event user.UserCreatedEvent) error {
	user := UserProfile{
		PK:     "USER#" + event.UserID,
		SK:     "PROFILE",
		UserID: event.UserID,
		Name:   event.Name,
	}

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{Item: item, TableName: &tableName})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

func handleRoomCreatedEvent(ctx context.Context, client *dynamodb.Client, event room.ChatRoomCreatedEvent) error {
	roomMetadata := RoomMetadata{
		PK:     "ROOM#" + event.ChatRoomID,
		SK:     "METADATA",
		RoomID: event.ChatRoomID,
		Name:   event.Name,
	}
	item, err := attributevalue.MarshalMap(roomMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal room: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{Item: item, TableName: &tableName})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

func handleMembershipCreatedEvent(ctx context.Context, client *dynamodb.Client, event membership.MembershipCreatedEvent) error {
	joinedRoom := JoinedRoom{
		PK:           "USER#" + event.UserId,
		SK:           "ROOM#" + event.ChatRoomId,
		MembershipID: event.Id,
		RoomID:       event.ChatRoomId,
		UserID:       event.UserId,
	}
	membership := Membership{
		PK:           "ROOM#" + event.ChatRoomId,
		SK:           "USER#" + event.UserId,
		MembershipID: event.Id,
		RoomID:       event.ChatRoomId,
		UserID:       event.UserId,
	}
	joinedRoomAv, err := attributevalue.MarshalMap(joinedRoom)
	if err != nil {
		return fmt.Errorf("failed to marshal joined room: %w", err)
	}
	membershipAv, err := attributevalue.MarshalMap(membership)
	if err != nil {
		return fmt.Errorf("failed to marshal membership: %w", err)
	}

	_, err = client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					Item:                                joinedRoomAv,
					TableName:                           &tableName,
					ConditionExpression:                 nil,
					ExpressionAttributeNames:            nil,
					ExpressionAttributeValues:           nil,
					ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureNone,
				},
				ConditionCheck: nil,
				Delete:         nil,
				Update:         nil,
			},
			{
				Put: &types.Put{
					Item:                                membershipAv,
					TableName:                           &tableName,
					ConditionExpression:                 nil,
					ExpressionAttributeNames:            nil,
					ExpressionAttributeValues:           nil,
					ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureNone,
				},
				ConditionCheck: nil,
				Delete:         nil,
				Update:         nil,
			},
		},
		ClientRequestToken:          nil,
		ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
		ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone,
	})
	if err != nil {
		return fmt.Errorf("failed to transact write items: %w", err)
	}

	return nil
}

func handleMessagePostedEvent(ctx context.Context, client *dynamodb.Client, event message.MessagePostedEvent) error {
	message := Message{
		PK:        "ROOM#" + event.ChatRoomID,
		SK:        "MESSAGE#" + event.ID,
		MessageID: event.ID,
		Body:      event.Body,
		RoomID:    event.ChatRoomID,
		UserID:    event.AuthorUserID,
	}
	messageAv, err := attributevalue.MarshalMap(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{Item: messageAv, TableName: &tableName})
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
		case message.MessagePostedEventType:
			devent, err := extractEventData[message.MessagePostedEvent](eb)
			if err != nil {
				return fmt.Errorf("failed to extract domain event: %w", err)
			}
			err = handleMessagePostedEvent(ctx, c.client, devent)
			if err != nil {
				return fmt.Errorf("failed to handle message posted event: %w", err)
			}
			slog.DebugContext(ctx, "message posted", "domain_data", devent)
		default:
			return fmt.Errorf("failed to handle event, unknown event type: event_type = %s", data.Type)
		}

	}

	return nil
}
