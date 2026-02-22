package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sudame/chat/internal/domain/user"
)

type mockUserRepository struct {
	saveFunc    func(ctx context.Context, u *user.User) error
	findByIDFunc func(ctx context.Context, id string) (*user.User, error)
}

func (m *mockUserRepository) Save(ctx context.Context, u *user.User) error {
	return m.saveFunc(ctx, u)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	return m.findByIDFunc(ctx, id)
}

func TestCreateUserService_Exec(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: ユーザーを作成してIDを返す", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			saveFunc: func(_ context.Context, _ *user.User) error {
				return nil
			},
		}
		svc := NewCreateUserService(mockRepo)

		userID, err := svc.Exec(ctx, "Alice")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if userID == nil || *userID == "" {
			t.Error("expected non-empty user ID")
		}
	})

	t.Run("異常系: 空の名前でエラーを返す", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			saveFunc: func(_ context.Context, _ *user.User) error {
				return nil
			},
		}
		svc := NewCreateUserService(mockRepo)

		_, err := svc.Exec(ctx, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, user.ErrEmptyName) {
			t.Errorf("expected ErrEmptyName to be wrapped, got %v", err)
		}
	})

	t.Run("異常系: リポジトリのSaveがエラーを返す場合エラーを返す", func(t *testing.T) {
		saveErr := errors.New("DB write failed")
		mockRepo := &mockUserRepository{
			saveFunc: func(_ context.Context, _ *user.User) error {
				return saveErr
			},
		}
		svc := NewCreateUserService(mockRepo)

		_, err := svc.Exec(ctx, "Bob")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, saveErr) {
			t.Errorf("expected saveErr to be wrapped, got %v", err)
		}
	})
}
