package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/internal/domain/user"
)

func TestCreateRoomService_Exec(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: ルームを作成してIDを返す", func(t *testing.T) {
		userRepo := &mockUserRepository{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return user.ReconstructUser("user:01", "Alice"), nil
			},
		}
		roomRepo := &mockRoomRepository{
			saveFunc: func(_ context.Context, _ *room.Room) error { return nil },
		}
		svc := NewCreateRoomService(userRepo, roomRepo)

		roomID, err := svc.Exec(ctx, "テストルーム", "user:01")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if roomID == nil || *roomID == "" {
			t.Error("expected non-empty room ID")
		}
	})

	t.Run("異常系: 存在しないユーザーの場合エラーを返す", func(t *testing.T) {
		userRepo := &mockUserRepository{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return nil, user.ErrNotFound
			},
		}
		roomRepo := &mockRoomRepository{
			saveFunc: func(_ context.Context, _ *room.Room) error { return nil },
		}
		svc := NewCreateRoomService(userRepo, roomRepo)

		_, err := svc.Exec(ctx, "テストルーム", "nonexistent-user")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, user.ErrNotFound) {
			t.Errorf("expected ErrNotFound to be wrapped, got %v", err)
		}
	})

	t.Run("異常系: 空のルーム名でエラーを返す", func(t *testing.T) {
		userRepo := &mockUserRepository{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return user.ReconstructUser("user:01", "Alice"), nil
			},
		}
		roomRepo := &mockRoomRepository{
			saveFunc: func(_ context.Context, _ *room.Room) error { return nil },
		}
		svc := NewCreateRoomService(userRepo, roomRepo)

		_, err := svc.Exec(ctx, "", "user:01")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("異常系: リポジトリのSaveがエラーを返す場合エラーを返す", func(t *testing.T) {
		saveErr := errors.New("DB write failed")
		userRepo := &mockUserRepository{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return user.ReconstructUser("user:01", "Alice"), nil
			},
		}
		roomRepo := &mockRoomRepository{
			saveFunc: func(_ context.Context, _ *room.Room) error { return saveErr },
		}
		svc := NewCreateRoomService(userRepo, roomRepo)

		_, err := svc.Exec(ctx, "テストルーム", "user:01")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, saveErr) {
			t.Errorf("expected saveErr to be wrapped, got %v", err)
		}
	})
}
