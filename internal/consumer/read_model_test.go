package consumer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/internal/domain/user"
	"github.com/sudame/chat/internal/ticdc"
)

// newMockDynamoDBClient は httptest.Server を使って DynamoDB エンドポイントをモックする。
// handler にはレスポンスを返す HTTP ハンドラーを渡す。
func newMockDynamoDBClient(handler http.Handler) (*dynamodb.Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	client := dynamodb.NewFromConfig(aws.Config{
		Region: "us-east-1",
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test",
				SecretAccessKey: "test",
				SessionToken:    "",
			}, nil
		}),
	}, func(o *dynamodb.Options) {
		o.BaseEndpoint = &srv.URL
	})
	return client, srv
}

// dynamoDBSuccessHandler は DynamoDB PutItem リクエストに対して空の成功レスポンスを返す。
var dynamoDBSuccessHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`)) //nolint:errcheck
})

func TestReadModelConsumer_Consume_MalformedPayload(t *testing.T) {
	client, srv := newMockDynamoDBClient(dynamoDBSuccessHandler)
	defer srv.Close()
	consumer := NewReadModelConsumer(client)

	event := ticdc.Event{
		Data: []ticdc.EventData{
			{
				ID:      "evt-1",
				Type:    room.ChatRoomCreatedEventType,
				Payload: json.RawMessage(`not-valid-json`),
			},
		},
	}

	err := consumer.Consume(context.Background(), event)
	if err == nil {
		t.Fatal("expected error for malformed payload, got nil")
	}
}

func TestReadModelConsumer_Consume_UnknownEventType(t *testing.T) {
	client, srv := newMockDynamoDBClient(dynamoDBSuccessHandler)
	defer srv.Close()
	consumer := NewReadModelConsumer(client)

	unknownEvent := map[string]string{"some": "data"}
	event := ticdc.Event{
		Data: []ticdc.EventData{
			{
				ID:      "evt-1",
				Type:    "unknown.event.type",
				Payload: buildPayload(unknownEvent),
			},
		},
	}

	err := consumer.Consume(context.Background(), event)
	if err == nil {
		t.Fatal("expected error for unknown event type, got nil")
	}
}

func TestReadModelConsumer_Consume_EmptyEvent(t *testing.T) {
	client, srv := newMockDynamoDBClient(dynamoDBSuccessHandler)
	defer srv.Close()
	consumer := NewReadModelConsumer(client)

	err := consumer.Consume(context.Background(), ticdc.Event{Data: []ticdc.EventData{}})
	if err != nil {
		t.Fatalf("expected no error for empty event, got %v", err)
	}
}

func TestReadModelConsumer_Consume_UserCreatedEvent(t *testing.T) {
	client, srv := newMockDynamoDBClient(dynamoDBSuccessHandler)
	defer srv.Close()
	consumer := NewReadModelConsumer(client)

	userEvent := user.UserCreatedEvent{UserID: "user:01", Name: "Alice"}
	event := ticdc.Event{
		Data: []ticdc.EventData{
			{
				ID:      "evt-1",
				Type:    user.UserCreatedEventType,
				Payload: buildPayload(userEvent),
			},
		},
	}

	err := consumer.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestReadModelConsumer_Consume_ChatRoomCreatedEvent(t *testing.T) {
	client, srv := newMockDynamoDBClient(dynamoDBSuccessHandler)
	defer srv.Close()
	consumer := NewReadModelConsumer(client)

	roomEvent := room.ChatRoomCreatedEvent{
		ChatRoomID:    "chat-room:01",
		Name:          "テストルーム",
		CreatorUserID: "user:01",
	}
	event := ticdc.Event{
		Data: []ticdc.EventData{
			{
				ID:      "evt-1",
				Type:    room.ChatRoomCreatedEventType,
				Payload: buildPayload(roomEvent),
			},
		},
	}

	err := consumer.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestReadModelConsumer_Consume_MembershipCreatedEvent(t *testing.T) {
	client, srv := newMockDynamoDBClient(dynamoDBSuccessHandler)
	defer srv.Close()
	consumer := NewReadModelConsumer(client)

	msEvent := membership.MembershipCreatedEvent{
		Id:         "membership-01",
		UserId:     "user:01",
		ChatRoomId: "chat-room:01",
	}
	event := ticdc.Event{
		Data: []ticdc.EventData{
			{
				ID:      "evt-1",
				Type:    membership.MembershipCreatedEventType,
				Payload: buildPayload(msEvent),
			},
		},
	}

	err := consumer.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestReadModelConsumer_Consume_DynamoDBError(t *testing.T) {
	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"__type":"InternalServerError","message":"internal server error"}`)) //nolint:errcheck
	})
	client, srv := newMockDynamoDBClient(errorHandler)
	defer srv.Close()
	consumer := NewReadModelConsumer(client)

	userEvent := user.UserCreatedEvent{UserID: "user:01", Name: "Alice"}
	event := ticdc.Event{
		Data: []ticdc.EventData{
			{
				ID:      "evt-1",
				Type:    user.UserCreatedEventType,
				Payload: buildPayload(userEvent),
			},
		},
	}

	err := consumer.Consume(context.Background(), event)
	if err == nil {
		t.Fatal("expected error from DynamoDB failure, got nil")
	}
}
