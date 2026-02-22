package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/internal/domain/user"
)

type mockRoomQuery struct {
	checkRoomExistenceFunc func(req room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error)
}

func (m *mockRoomQuery) CheckRoomExistence(req room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
	return m.checkRoomExistenceFunc(req)
}

type mockMembershipRepository struct {
	saveFunc    func(ctx context.Context, m *membership.Membership) error
	findByIdFunc func(ctx context.Context, id string) (*membership.Membership, error)
}

func (m *mockMembershipRepository) Save(ctx context.Context, ms *membership.Membership) error {
	return m.saveFunc(ctx, ms)
}

func (m *mockMembershipRepository) FindById(ctx context.Context, id string) (*membership.Membership, error) {
	return m.findByIdFunc(ctx, id)
}

func TestCreateMembershipService_Exec(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: メンバーシップを作成できる", func(t *testing.T) {
		userRepo := &mockUserRepository{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return user.ReconstructUser("user:01", "Alice"), nil
			},
		}
		roomQuery := &mockRoomQuery{
			checkRoomExistenceFunc: func(_ room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
				return room.CheckRoomExistenceResponse{Existence: true}, nil
			},
		}
		membershipRepo := &mockMembershipRepository{
			saveFunc: func(_ context.Context, _ *membership.Membership) error { return nil },
		}
		svc := NewCreateMembershipService(userRepo, roomQuery, membershipRepo)

		err := svc.Exec(ctx, "user:01", "chat-room:01")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("異常系: 存在しないユーザーの場合エラーを返す", func(t *testing.T) {
		userRepo := &mockUserRepository{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return nil, user.ErrNotFound
			},
		}
		roomQuery := &mockRoomQuery{
			checkRoomExistenceFunc: func(_ room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
				return room.CheckRoomExistenceResponse{Existence: true}, nil
			},
		}
		membershipRepo := &mockMembershipRepository{
			saveFunc: func(_ context.Context, _ *membership.Membership) error { return nil },
		}
		svc := NewCreateMembershipService(userRepo, roomQuery, membershipRepo)

		err := svc.Exec(ctx, "nonexistent-user", "chat-room:01")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, user.ErrNotFound) {
			t.Errorf("expected ErrNotFound to be wrapped, got %v", err)
		}
	})

	t.Run("異常系: ルームが存在しない場合エラーを返す", func(t *testing.T) {
		userRepo := &mockUserRepository{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return user.ReconstructUser("user:01", "Alice"), nil
			},
		}
		roomQuery := &mockRoomQuery{
			checkRoomExistenceFunc: func(_ room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
				return room.CheckRoomExistenceResponse{Existence: false}, nil
			},
		}
		membershipRepo := &mockMembershipRepository{
			saveFunc: func(_ context.Context, _ *membership.Membership) error { return nil },
		}
		svc := NewCreateMembershipService(userRepo, roomQuery, membershipRepo)

		err := svc.Exec(ctx, "user:01", "nonexistent-room")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("異常系: ルームサービスへの問い合わせが失敗した場合エラーを返す", func(t *testing.T) {
		queryErr := errors.New("room service unavailable")
		userRepo := &mockUserRepository{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return user.ReconstructUser("user:01", "Alice"), nil
			},
		}
		roomQuery := &mockRoomQuery{
			checkRoomExistenceFunc: func(_ room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
				return room.CheckRoomExistenceResponse{}, queryErr
			},
		}
		membershipRepo := &mockMembershipRepository{
			saveFunc: func(_ context.Context, _ *membership.Membership) error { return nil },
		}
		svc := NewCreateMembershipService(userRepo, roomQuery, membershipRepo)

		err := svc.Exec(ctx, "user:01", "chat-room:01")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, queryErr) {
			t.Errorf("expected queryErr to be wrapped, got %v", err)
		}
	})

	t.Run("異常系: メンバーシップのSaveがエラーを返す場合エラーを返す", func(t *testing.T) {
		saveErr := errors.New("DB write failed")
		userRepo := &mockUserRepository{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return user.ReconstructUser("user:01", "Alice"), nil
			},
		}
		roomQuery := &mockRoomQuery{
			checkRoomExistenceFunc: func(_ room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
				return room.CheckRoomExistenceResponse{Existence: true}, nil
			},
		}
		membershipRepo := &mockMembershipRepository{
			saveFunc: func(_ context.Context, _ *membership.Membership) error { return saveErr },
		}
		svc := NewCreateMembershipService(userRepo, roomQuery, membershipRepo)

		err := svc.Exec(ctx, "user:01", "chat-room:01")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, saveErr) {
			t.Errorf("expected saveErr to be wrapped, got %v", err)
		}
	})
}
