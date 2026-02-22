package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sudame/chat/internal/domain/room"
)

type mockRoomRepository struct {
	findByIDFunc func(ctx context.Context, id string) (*room.Room, error)
	saveFunc     func(ctx context.Context, r *room.Room) error
}

func (m *mockRoomRepository) FindByID(ctx context.Context, id string) (*room.Room, error) {
	return m.findByIDFunc(ctx, id)
}

func (m *mockRoomRepository) Save(ctx context.Context, r *room.Room) error {
	return m.saveFunc(ctx, r)
}

func TestCheckRoomExistenceService_Exec(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: ルームが存在する場合 Existence=true を返す", func(t *testing.T) {
		mockRepo := &mockRoomRepository{
			findByIDFunc: func(_ context.Context, _ string) (*room.Room, error) {
				return room.ReconstructRoom("chat-room:01", "テストルーム"), nil
			},
		}
		svc := NewCheckRoomExistenceService(mockRepo)

		result, err := svc.Exec(ctx, "chat-room:01")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !result.Existence {
			t.Error("expected Existence=true, got false")
		}
	})

	t.Run("正常系: ルームが存在しない場合 Existence=false を返す", func(t *testing.T) {
		mockRepo := &mockRoomRepository{
			findByIDFunc: func(_ context.Context, _ string) (*room.Room, error) {
				return nil, room.ErrNotFound
			},
		}
		svc := NewCheckRoomExistenceService(mockRepo)

		result, err := svc.Exec(ctx, "chat-room:nonexistent")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Existence {
			t.Error("expected Existence=false, got true")
		}
	})

	t.Run("異常系: リポジトリが予期しないエラーを返す場合エラーを返す", func(t *testing.T) {
		dbErr := errors.New("database connection failed")
		mockRepo := &mockRoomRepository{
			findByIDFunc: func(_ context.Context, _ string) (*room.Room, error) {
				return nil, dbErr
			},
		}
		svc := NewCheckRoomExistenceService(mockRepo)

		_, err := svc.Exec(ctx, "chat-room:01")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, dbErr) {
			t.Errorf("expected error to wrap dbErr, got %v", err)
		}
	})
}
