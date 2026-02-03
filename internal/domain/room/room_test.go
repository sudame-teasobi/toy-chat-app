package room

import (
	"testing"
)

func TestNewChatRoom(t *testing.T) {
	t.Run("正常系: 有効な名前でチャットルームを作成できる", func(t *testing.T) {
		cr, err := NewRoom("テストルーム", "user-100")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cr == nil {
			t.Fatal("expected chatroom to be non-nil")
		}
		if cr.ID() == "" {
			t.Error("expected ID to be non-empty")
		}
		if cr.Name() != "テストルーム" {
			t.Errorf("expected name 'テストルーム', got '%s'", cr.Name())
		}
	})

	t.Run("異常系: 空の名前でエラーを返す", func(t *testing.T) {
		cr, err := NewRoom("", "user-100")
		if err != ErrEmptyName {
			t.Errorf("expected ErrEmptyName, got %v", err)
		}
		if cr != nil {
			t.Error("expected chatroom to be nil")
		}
	})

	t.Run("ChatRoomCreatedEventのみが記録される", func(t *testing.T) {
		cr, _ := NewRoom("テストルーム", "user-100")
		events := cr.Events()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		createdEvent, ok := events[0].(*ChatRoomCreatedEvent)
		if !ok {
			t.Fatal("expected first event to be ChatRoomCreatedEvent")
		}
		if createdEvent.ChatRoomID != cr.ID() {
			t.Errorf("expected ChatRoomID '%s', got '%s'", cr.ID(), createdEvent.ChatRoomID)
		}
		if createdEvent.Name != "テストルーム" {
			t.Errorf("expected Name 'テストルーム', got '%s'", createdEvent.Name)
		}
		if createdEvent.CreatorUserID != "user-100" {
			t.Errorf("expected CreatorUserID 'user-100', got '%s'", createdEvent.CreatorUserID)
		}
	})
}

func TestReconstructChatRoom(t *testing.T) {
	t.Run("永続化からチャットルームを再構築できる", func(t *testing.T) {
		cr := ReconstructRoom("chatroom-42", "再構築ルーム")
		if cr == nil {
			t.Fatal("expected chatroom to be non-nil")
		}
		if cr.ID() != "chatroom-42" {
			t.Errorf("expected ID 'chatroom-42', got '%s'", cr.ID())
		}
		if cr.Name() != "再構築ルーム" {
			t.Errorf("expected name '再構築ルーム', got '%s'", cr.Name())
		}
	})

	t.Run("再構築時はイベントが記録されない", func(t *testing.T) {
		cr := ReconstructRoom("chatroom-1", "テスト")
		if len(cr.Events()) != 0 {
			t.Errorf("expected 0 events, got %d", len(cr.Events()))
		}
	})
}

func TestChatRoom_ClearEvents(t *testing.T) {
	t.Run("イベントがクリアされる", func(t *testing.T) {
		cr, _ := NewRoom("テストルーム", "user-100")
		if len(cr.Events()) == 0 {
			t.Fatal("expected events before clear")
		}

		cr.ClearEvents()

		if len(cr.Events()) != 0 {
			t.Errorf("expected 0 events after clear, got %d", len(cr.Events()))
		}
	})
}
