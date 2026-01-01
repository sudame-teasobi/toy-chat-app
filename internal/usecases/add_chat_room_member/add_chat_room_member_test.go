package addchatroommember_test

import (
	"context"
	"testing"

	"github.com/sudame/chat/internal/events"
	"github.com/sudame/chat/internal/models"
	addchatroommember "github.com/sudame/chat/internal/usecases/add_chat_room_member"
)

// モックリポジトリ
type mockRepository struct {
	userExists     bool
	chatRoomExists bool
	isMember       bool
	addedMember    *models.ChatRoomMember
	savedEvents    []events.Event
}

func (m *mockRepository) UserExists(ctx context.Context, userID int64) (bool, error) {
	return m.userExists, nil
}

func (m *mockRepository) ChatRoomExists(ctx context.Context, chatRoomID int64) (bool, error) {
	return m.chatRoomExists, nil
}

func (m *mockRepository) IsMember(ctx context.Context, chatRoomID, userID int64) (bool, error) {
	return m.isMember, nil
}

func (m *mockRepository) AddMember(ctx context.Context, chatRoomID, userID int64) (*models.ChatRoomMember, error) {
	m.addedMember = &models.ChatRoomMember{
		ID:         1,
		ChatRoomID: chatRoomID,
		UserID:     userID,
	}
	return m.addedMember, nil
}

func (m *mockRepository) SaveEvent(ctx context.Context, event events.Event) error {
	m.savedEvents = append(m.savedEvents, event)
	return nil
}

func TestAddChatRoomMember_正常にメンバーを追加できる(t *testing.T) {
	// Arrange
	repo := &mockRepository{
		userExists:     true,
		chatRoomExists: true,
		isMember:       false,
	}
	usecase := addchatroommember.NewUsecase(repo)
	input := addchatroommember.Input{
		ChatRoomID: 1,
		UserID:     42,
	}

	// Act
	output, err := usecase.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if output.Member == nil {
		t.Fatal("メンバーが返されていません")
	}
	if output.Member.ChatRoomID != 1 {
		t.Errorf("ChatRoomIDが一致しません: got %d, want %d", output.Member.ChatRoomID, 1)
	}
	if output.Member.UserID != 42 {
		t.Errorf("UserIDが一致しません: got %d, want %d", output.Member.UserID, 42)
	}
}

func TestAddChatRoomMember_MemberAddedイベントが発行される(t *testing.T) {
	// Arrange
	repo := &mockRepository{
		userExists:     true,
		chatRoomExists: true,
		isMember:       false,
	}
	usecase := addchatroommember.NewUsecase(repo)
	input := addchatroommember.Input{
		ChatRoomID: 1,
		UserID:     42,
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	// MemberAddedイベントが発行されているか確認
	var foundEvent *events.MemberAddedEvent
	for _, e := range repo.savedEvents {
		if evt, ok := e.(*events.MemberAddedEvent); ok {
			foundEvent = evt
			break
		}
	}
	if foundEvent == nil {
		t.Fatal("MemberAddedイベントが発行されていません")
	}
	if foundEvent.ChatRoomID != 1 {
		t.Errorf("イベントのChatRoomIDが一致しません: got %d, want %d", foundEvent.ChatRoomID, 1)
	}
	if foundEvent.UserID != 42 {
		t.Errorf("イベントのUserIDが一致しません: got %d, want %d", foundEvent.UserID, 42)
	}
}

func TestAddChatRoomMember_存在しないユーザーを追加できない(t *testing.T) {
	// Arrange
	repo := &mockRepository{
		userExists:     false,
		chatRoomExists: true,
		isMember:       false,
	}
	usecase := addchatroommember.NewUsecase(repo)
	input := addchatroommember.Input{
		ChatRoomID: 1,
		UserID:     999,
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("エラーが発生するはずですが、発生しませんでした")
	}
	if err != addchatroommember.ErrUserNotFound {
		t.Errorf("エラーが一致しません: got %v, want %v", err, addchatroommember.ErrUserNotFound)
	}
}

func TestAddChatRoomMember_存在しないチャットルームに追加できない(t *testing.T) {
	// Arrange
	repo := &mockRepository{
		userExists:     true,
		chatRoomExists: false,
		isMember:       false,
	}
	usecase := addchatroommember.NewUsecase(repo)
	input := addchatroommember.Input{
		ChatRoomID: 999,
		UserID:     1,
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("エラーが発生するはずですが、発生しませんでした")
	}
	if err != addchatroommember.ErrChatRoomNotFound {
		t.Errorf("エラーが一致しません: got %v, want %v", err, addchatroommember.ErrChatRoomNotFound)
	}
}

func TestAddChatRoomMember_すでにメンバーの場合は追加できない(t *testing.T) {
	// Arrange
	repo := &mockRepository{
		userExists:     true,
		chatRoomExists: true,
		isMember:       true,
	}
	usecase := addchatroommember.NewUsecase(repo)
	input := addchatroommember.Input{
		ChatRoomID: 1,
		UserID:     42,
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("エラーが発生するはずですが、発生しませんでした")
	}
	if err != addchatroommember.ErrAlreadyMember {
		t.Errorf("エラーが一致しません: got %v, want %v", err, addchatroommember.ErrAlreadyMember)
	}
}
