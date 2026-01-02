package user

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	t.Run("正常系: 有効な名前でユーザーを作成できる", func(t *testing.T) {
		usr, err := NewUser(1, "テストユーザー")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if usr == nil {
			t.Fatal("expected user to be non-nil")
		}
		if usr.ID() != 1 {
			t.Errorf("expected ID 1, got %d", usr.ID())
		}
		if usr.Name() != "テストユーザー" {
			t.Errorf("expected name 'テストユーザー', got '%s'", usr.Name())
		}
	})

	t.Run("異常系: 空の名前でエラーを返す", func(t *testing.T) {
		usr, err := NewUser(1, "")
		if err != ErrEmptyName {
			t.Errorf("expected ErrEmptyName, got %v", err)
		}
		if usr != nil {
			t.Error("expected user to be nil")
		}
	})

	t.Run("メンバー追加時にイベントが記録される", func(t *testing.T) {
		usr, _ := NewUser(1, "テストユーザー")

		events := usr.Events()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		userCreatedEvent, ok := events[0].(*UserCreatedEvent)
		if !ok {
			t.Fatal("expected UserCreatedEvent")
		}
		if userCreatedEvent.UserID != 1 {
			t.Errorf("expected UserID 1, got %d", userCreatedEvent.UserID)
		}
		if userCreatedEvent.Name != "テストユーザー" {
			t.Errorf("expected Name 'テストユーザー', got '%s'", userCreatedEvent.Name)
		}
	})
}

func TestReconstructUser(t *testing.T) {
	t.Run("永続化からユーザーを再構築できる", func(t *testing.T) {
		usr := ReconstructUser(42, "再構築ユーザー")
		if usr == nil {
			t.Fatal("expected user to be non-nil")
		}
		if usr.ID() != 42 {
			t.Errorf("expected ID 42, got %d", usr.ID())
		}
		if usr.Name() != "再構築ユーザー" {
			t.Errorf("expected name '再構築ユーザー', got '%s'", usr.Name())
		}
	})
}
