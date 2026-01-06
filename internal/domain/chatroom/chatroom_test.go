package chatroom

import (
	"testing"
)

func TestNewChatRoom(t *testing.T) {
	t.Run("正常系: 有効な名前でチャットルームを作成できる", func(t *testing.T) {
		cr, err := NewChatRoom("テストルーム", "user-100")
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
		cr, err := NewChatRoom("", "user-100")
		if err != ErrEmptyName {
			t.Errorf("expected ErrEmptyName, got %v", err)
		}
		if cr != nil {
			t.Error("expected chatroom to be nil")
		}
	})

	t.Run("作成時はメンバーが空（非同期で追加される）", func(t *testing.T) {
		cr, _ := NewChatRoom("テストルーム", "user-100")
		members := cr.Members()
		if len(members) != 0 {
			t.Errorf("expected 0 members, got %d", len(members))
		}
	})

	t.Run("ChatRoomCreatedEventのみが記録される", func(t *testing.T) {
		cr, _ := NewChatRoom("テストルーム", "user-100")
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
		members := []Member{
			ReconstructMember("member-1", "user-100"),
			ReconstructMember("member-2", "user-200"),
		}
		cr := ReconstructChatRoom("chatroom-42", "再構築ルーム", members)
		if cr == nil {
			t.Fatal("expected chatroom to be non-nil")
		}
		if cr.ID() != "chatroom-42" {
			t.Errorf("expected ID 'chatroom-42', got '%s'", cr.ID())
		}
		if cr.Name() != "再構築ルーム" {
			t.Errorf("expected name '再構築ルーム', got '%s'", cr.Name())
		}
		if len(cr.Members()) != 2 {
			t.Errorf("expected 2 members, got %d", len(cr.Members()))
		}
	})

	t.Run("再構築時はイベントが記録されない", func(t *testing.T) {
		cr := ReconstructChatRoom("chatroom-1", "テスト", nil)
		if len(cr.Events()) != 0 {
			t.Errorf("expected 0 events, got %d", len(cr.Events()))
		}
	})
}

func TestChatRoom_AddMember(t *testing.T) {
	t.Run("正常系: 新しいメンバーを追加できる", func(t *testing.T) {
		cr, _ := NewChatRoom("テストルーム", "user-100")
		cr.ClearEvents()

		err := cr.AddMember("user-200")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !cr.IsMember("user-200") {
			t.Error("expected new user to be a member")
		}
		if len(cr.Members()) != 1 {
			t.Errorf("expected 1 member, got %d", len(cr.Members()))
		}
	})

	t.Run("異常系: 既存メンバーを追加するとエラー", func(t *testing.T) {
		cr, _ := NewChatRoom("テストルーム", "user-100")
		_ = cr.AddMember("user-100") // まずメンバーを追加

		err := cr.AddMember("user-100") // 再度追加しようとする
		if err != ErrAlreadyMember {
			t.Errorf("expected ErrAlreadyMember, got %v", err)
		}
	})

	t.Run("メンバー追加時にイベントが記録される", func(t *testing.T) {
		cr, _ := NewChatRoom("テストルーム", "user-100")
		cr.ClearEvents()

		_ = cr.AddMember("user-200")

		events := cr.Events()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		memberAddedEvent, ok := events[0].(*MemberAddedEvent)
		if !ok {
			t.Fatal("expected MemberAddedEvent")
		}
		if memberAddedEvent.UserID != "user-200" {
			t.Errorf("expected UserID 'user-200', got '%s'", memberAddedEvent.UserID)
		}
	})
}

func TestChatRoom_IsMember(t *testing.T) {
	t.Run("メンバーの場合trueを返す", func(t *testing.T) {
		cr, _ := NewChatRoom("テストルーム", "user-100")
		_ = cr.AddMember("user-100") // メンバーを追加
		if !cr.IsMember("user-100") {
			t.Error("expected true for member")
		}
	})

	t.Run("非メンバーの場合falseを返す", func(t *testing.T) {
		cr, _ := NewChatRoom("テストルーム", "user-100")
		if cr.IsMember("user-999") {
			t.Error("expected false for non-member")
		}
	})
}

func TestChatRoom_ClearEvents(t *testing.T) {
	t.Run("イベントがクリアされる", func(t *testing.T) {
		cr, _ := NewChatRoom("テストルーム", "user-100")
		if len(cr.Events()) == 0 {
			t.Fatal("expected events before clear")
		}

		cr.ClearEvents()

		if len(cr.Events()) != 0 {
			t.Errorf("expected 0 events after clear, got %d", len(cr.Events()))
		}
	})
}

func TestChatRoom_Members_ReturnsCopy(t *testing.T) {
	t.Run("Membersは防御的コピーを返す", func(t *testing.T) {
		cr, _ := NewChatRoom("テストルーム", "user-100")
		_ = cr.AddMember("user-100") // メンバーを追加
		members1 := cr.Members()
		members2 := cr.Members()

		if &members1[0] == &members2[0] {
			t.Error("expected Members() to return a copy")
		}
	})
}
