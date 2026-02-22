package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/internal/domain/user"
	"github.com/sudame/chat/internal/service"
	"github.com/sudame/chat/internal/ticdc"
)

// ---- Mock implementations ----

type mockUserRepo struct {
	saveFunc     func(ctx context.Context, u *user.User) error
	findByIDFunc func(ctx context.Context, id string) (*user.User, error)
}

func (m *mockUserRepo) Save(ctx context.Context, u *user.User) error {
	return m.saveFunc(ctx, u)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*user.User, error) {
	return m.findByIDFunc(ctx, id)
}

type mockRoomQuery struct {
	checkRoomExistenceFunc func(req room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error)
}

func (m *mockRoomQuery) CheckRoomExistence(req room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
	return m.checkRoomExistenceFunc(req)
}

type mockMembershipRepo struct {
	saveFunc     func(ctx context.Context, m *membership.Membership) error
	findByIdFunc func(ctx context.Context, id string) (*membership.Membership, error)
}

func (m *mockMembershipRepo) Save(ctx context.Context, ms *membership.Membership) error {
	return m.saveFunc(ctx, ms)
}
func (m *mockMembershipRepo) FindById(ctx context.Context, id string) (*membership.Membership, error) {
	return m.findByIdFunc(ctx, id)
}

// ---- Helper ----

// buildPayload wraps a domain event struct as double-encoded JSON (string containing JSON),
// matching how TiCDC encodes payloads.
func buildPayload(v any) json.RawMessage {
	inner, _ := json.Marshal(v)
	outer, _ := json.Marshal(string(inner))
	return outer
}

func newMembershipConsumer(
	userRepo user.Repository,
	roomQuery room.Query,
	membershipRepo membership.Repository,
) *MembershipConsumer {
	svc := service.NewCreateMembershipService(userRepo, roomQuery, membershipRepo)
	return NewMembershipConsumer(svc)
}

// ---- Tests ----

func TestMembershipConsumer_Consume_SkipsNonMatchingEventType(t *testing.T) {
	// このテストは Issue #5 の根本原因を検証する:
	// ChatRoomCreatedEvent 以外のイベントタイプを受け取ったときに
	// return nil ではなく continue で次のイベントに進むことを確認する。

	execCallCount := 0
	userRepo := &mockUserRepo{
		findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
			return user.ReconstructUser("user:01", "Alice"), nil
		},
	}
	roomQuery := &mockRoomQuery{
		checkRoomExistenceFunc: func(_ room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
			return room.CheckRoomExistenceResponse{Existence: true}, nil
		},
	}
	membershipRepo := &mockMembershipRepo{
		saveFunc: func(_ context.Context, _ *membership.Membership) error {
			execCallCount++
			return nil
		},
	}
	consumer := newMembershipConsumer(userRepo, roomQuery, membershipRepo)

	chatRoomEvent := room.ChatRoomCreatedEvent{
		ChatRoomID:    "chat-room:01",
		Name:          "テストルーム",
		CreatorUserID: "user:01",
	}

	event := ticdc.Event{
		Data: []ticdc.EventData{
			// 1つ目: ChatRoomCreatedEvent 以外 → スキップされるべき
			{
				ID:      "evt-1",
				Type:    "user.created",
				Payload: buildPayload(map[string]string{"user_id": "user:01", "name": "Alice"}),
			},
			// 2つ目: ChatRoomCreatedEvent → 処理されるべき
			{
				ID:      "evt-2",
				Type:    room.ChatRoomCreatedEventType,
				Payload: buildPayload(chatRoomEvent),
			},
		},
	}

	err := consumer.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// return nil だと 2つ目のイベントは処理されず execCallCount=0 になる。
	// continue だと 2つ目のイベントが処理されて execCallCount=1 になる。
	if execCallCount != 1 {
		t.Errorf(
			"expected service.Exec to be called once (continue behavior), but was called %d times"+
				" — if 0, the consumer used 'return nil' instead of 'continue'",
			execCallCount,
		)
	}
}

func TestMembershipConsumer_Consume_ProcessesChatRoomCreatedEvent(t *testing.T) {
	var capturedUserID, capturedRoomID string
	userRepo := &mockUserRepo{
		findByIDFunc: func(_ context.Context, id string) (*user.User, error) {
			return user.ReconstructUser(id, "Alice"), nil
		},
	}
	roomQuery := &mockRoomQuery{
		checkRoomExistenceFunc: func(req room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
			capturedRoomID = req.RoomID
			return room.CheckRoomExistenceResponse{Existence: true}, nil
		},
	}
	membershipRepo := &mockMembershipRepo{
		saveFunc: func(_ context.Context, m *membership.Membership) error {
			capturedUserID = m.UserId
			return nil
		},
	}
	consumer := newMembershipConsumer(userRepo, roomQuery, membershipRepo)

	chatRoomEvent := room.ChatRoomCreatedEvent{
		ChatRoomID:    "chat-room:01",
		Name:          "テストルーム",
		CreatorUserID: "user:01",
	}
	event := ticdc.Event{
		Data: []ticdc.EventData{
			{
				ID:      "evt-1",
				Type:    room.ChatRoomCreatedEventType,
				Payload: buildPayload(chatRoomEvent),
			},
		},
	}

	err := consumer.Consume(context.Background(), event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedUserID != "user:01" {
		t.Errorf("expected userID=user:01, got %q", capturedUserID)
	}
	if capturedRoomID != "chat-room:01" {
		t.Errorf("expected roomID=chat-room:01, got %q", capturedRoomID)
	}
}

func TestMembershipConsumer_Consume_ErrorOnMalformedPayload(t *testing.T) {
	userRepo := &mockUserRepo{
		findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
			return user.ReconstructUser("user:01", "Alice"), nil
		},
	}
	roomQuery := &mockRoomQuery{
		checkRoomExistenceFunc: func(_ room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
			return room.CheckRoomExistenceResponse{Existence: true}, nil
		},
	}
	membershipRepo := &mockMembershipRepo{
		saveFunc: func(_ context.Context, _ *membership.Membership) error { return nil },
	}
	consumer := newMembershipConsumer(userRepo, roomQuery, membershipRepo)

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

func TestMembershipConsumer_Consume_ErrorOnServiceFailure(t *testing.T) {
	serviceErr := errors.New("membership service error")
	userRepo := &mockUserRepo{
		findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
			return nil, serviceErr
		},
	}
	roomQuery := &mockRoomQuery{
		checkRoomExistenceFunc: func(_ room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
			return room.CheckRoomExistenceResponse{Existence: true}, nil
		},
	}
	membershipRepo := &mockMembershipRepo{
		saveFunc: func(_ context.Context, _ *membership.Membership) error { return nil },
	}
	consumer := newMembershipConsumer(userRepo, roomQuery, membershipRepo)

	chatRoomEvent := room.ChatRoomCreatedEvent{
		ChatRoomID:    "chat-room:01",
		Name:          "テストルーム",
		CreatorUserID: "user:01",
	}
	event := ticdc.Event{
		Data: []ticdc.EventData{
			{
				ID:      "evt-1",
				Type:    room.ChatRoomCreatedEventType,
				Payload: buildPayload(chatRoomEvent),
			},
		},
	}

	err := consumer.Consume(context.Background(), event)
	if err == nil {
		t.Fatal("expected error from service failure, got nil")
	}
	if !errors.Is(err, serviceErr) {
		t.Errorf("expected serviceErr to be wrapped, got %v", err)
	}
}

func TestMembershipConsumer_Consume_EmptyEvent(t *testing.T) {
	consumer := newMembershipConsumer(
		&mockUserRepo{},
		&mockRoomQuery{},
		&mockMembershipRepo{},
	)

	err := consumer.Consume(context.Background(), ticdc.Event{Data: []ticdc.EventData{}})
	if err != nil {
		t.Fatalf("expected no error for empty event, got %v", err)
	}
}
