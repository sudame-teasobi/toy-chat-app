package chatroom

import (
	"testing"
)

func TestNewChatRoom(t *testing.T) {
	t.Run("正常系: 有効な名前でチャットルームを作成できる", func(t *testing.T) {
		cr, err := NewChatRoom(1, "テストルーム", 100)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cr == nil {
			t.Fatal("expected chatroom to be non-nil")
		}
		if cr.ID() != 1 {
			t.Errorf("expected ID 1, got %d", cr.ID())
		}
		if cr.Name() != "テストルーム" {
			t.Errorf("expected name 'テストルーム', got '%s'", cr.Name())
		}
	})

	t.Run("異常系: 空の名前でエラーを返す", func(t *testing.T) {
		cr, err := NewChatRoom(1, "", 100)
		if err != ErrEmptyName {
			t.Errorf("expected ErrEmptyName, got %v", err)
		}
		if cr != nil {
			t.Error("expected chatroom to be nil")
		}
	})

	t.Run("作成者がメンバーとして追加される", func(t *testing.T) {
		cr, _ := NewChatRoom(1, "テストルーム", 100)
		if !cr.IsMember(100) {
			t.Error("expected creator to be a member")
		}
		members := cr.Members()
		if len(members) != 1 {
			t.Errorf("expected 1 member, got %d", len(members))
		}
		if members[0].UserID() != 100 {
			t.Errorf("expected member userID 100, got %d", members[0].UserID())
		}
	})

	t.Run("イベントが正しく記録される", func(t *testing.T) {
		cr, _ := NewChatRoom(1, "テストルーム", 100)
		events := cr.Events()
		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}

		createdEvent, ok := events[0].(*ChatRoomCreatedEvent)
		if !ok {
			t.Fatal("expected first event to be ChatRoomCreatedEvent")
		}
		if createdEvent.ChatRoomID != 1 {
			t.Errorf("expected ChatRoomID 1, got %d", createdEvent.ChatRoomID)
		}
		if createdEvent.Name != "テストルーム" {
			t.Errorf("expected Name 'テストルーム', got '%s'", createdEvent.Name)
		}

		memberAddedEvent, ok := events[1].(*MemberAddedEvent)
		if !ok {
			t.Fatal("expected second event to be MemberAddedEvent")
		}
		if memberAddedEvent.ChatRoomID != 1 {
			t.Errorf("expected ChatRoomID 1, got %d", memberAddedEvent.ChatRoomID)
		}
		if memberAddedEvent.UserID != 100 {
			t.Errorf("expected UserID 100, got %d", memberAddedEvent.UserID)
		}
	})
}

func TestReconstructChatRoom(t *testing.T) {
	t.Run("永続化からチャットルームを再構築できる", func(t *testing.T) {
		members := []Member{
			ReconstructMember(1, 100),
			ReconstructMember(2, 200),
		}
		cr := ReconstructChatRoom(42, "再構築ルーム", members)
		if cr == nil {
			t.Fatal("expected chatroom to be non-nil")
		}
		if cr.ID() != 42 {
			t.Errorf("expected ID 42, got %d", cr.ID())
		}
		if cr.Name() != "再構築ルーム" {
			t.Errorf("expected name '再構築ルーム', got '%s'", cr.Name())
		}
		if len(cr.Members()) != 2 {
			t.Errorf("expected 2 members, got %d", len(cr.Members()))
		}
	})

	t.Run("再構築時はイベントが記録されない", func(t *testing.T) {
		cr := ReconstructChatRoom(1, "テスト", nil)
		if len(cr.Events()) != 0 {
			t.Errorf("expected 0 events, got %d", len(cr.Events()))
		}
	})
}

func TestChatRoom_AddMember(t *testing.T) {
	t.Run("正常系: 新しいメンバーを追加できる", func(t *testing.T) {
		cr, _ := NewChatRoom(1, "テストルーム", 100)
		cr.ClearEvents()

		err := cr.AddMember(200)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !cr.IsMember(200) {
			t.Error("expected new user to be a member")
		}
		if len(cr.Members()) != 2 {
			t.Errorf("expected 2 members, got %d", len(cr.Members()))
		}
	})

	t.Run("異常系: 既存メンバーを追加するとエラー", func(t *testing.T) {
		cr, _ := NewChatRoom(1, "テストルーム", 100)

		err := cr.AddMember(100)
		if err != ErrAlreadyMember {
			t.Errorf("expected ErrAlreadyMember, got %v", err)
		}
	})

	t.Run("メンバー追加時にイベントが記録される", func(t *testing.T) {
		cr, _ := NewChatRoom(1, "テストルーム", 100)
		cr.ClearEvents()

		_ = cr.AddMember(200)

		events := cr.Events()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		memberAddedEvent, ok := events[0].(*MemberAddedEvent)
		if !ok {
			t.Fatal("expected MemberAddedEvent")
		}
		if memberAddedEvent.UserID != 200 {
			t.Errorf("expected UserID 200, got %d", memberAddedEvent.UserID)
		}
	})
}

func TestChatRoom_IsMember(t *testing.T) {
	t.Run("メンバーの場合trueを返す", func(t *testing.T) {
		cr, _ := NewChatRoom(1, "テストルーム", 100)
		if !cr.IsMember(100) {
			t.Error("expected true for member")
		}
	})

	t.Run("非メンバーの場合falseを返す", func(t *testing.T) {
		cr, _ := NewChatRoom(1, "テストルーム", 100)
		if cr.IsMember(999) {
			t.Error("expected false for non-member")
		}
	})
}

func TestChatRoom_ClearEvents(t *testing.T) {
	t.Run("イベントがクリアされる", func(t *testing.T) {
		cr, _ := NewChatRoom(1, "テストルーム", 100)
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
		cr, _ := NewChatRoom(1, "テストルーム", 100)
		members1 := cr.Members()
		members2 := cr.Members()

		if &members1[0] == &members2[0] {
			t.Error("expected Members() to return a copy")
		}
	})
}
