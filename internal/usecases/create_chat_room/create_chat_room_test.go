package createchatroom_test

import (
	"context"
	"testing"

	"github.com/sudame/chat/internal/domain/chatroom"
	createchatroom "github.com/sudame/chat/internal/usecases/create_chat_room"
)

// モックリポジトリ
type mockChatRoomRepository struct {
	savedRoom *chatroom.ChatRoom
}

func (m *mockChatRoomRepository) Save(ctx context.Context, room *chatroom.ChatRoom) error {
	m.savedRoom = room
	return nil
}

func (m *mockChatRoomRepository) FindByID(ctx context.Context, id int64) (*chatroom.ChatRoom, error) {
	return nil, nil
}

type mockUserRepository struct {
	userExists bool
}

func (m *mockUserRepository) UserExists(ctx context.Context, userID int64) (bool, error) {
	return m.userExists, nil
}

func TestCreateChatRoom_正常にチャットルームを作成できる(t *testing.T) {
	// Arrange
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{userExists: true}
	usecase := createchatroom.NewUsecase(chatRoomRepo, userRepo)
	input := createchatroom.Input{
		Name:      "テストルーム",
		CreatorID: 1,
	}

	// Act
	output, err := usecase.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if output.ChatRoom.Name() != "テストルーム" {
		t.Errorf("ルーム名が一致しません: got %s, want %s", output.ChatRoom.Name(), "テストルーム")
	}
}

func TestCreateChatRoom_作成者が自動的にメンバーとして追加される(t *testing.T) {
	// Arrange
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{userExists: true}
	usecase := createchatroom.NewUsecase(chatRoomRepo, userRepo)
	input := createchatroom.Input{
		Name:      "テストルーム",
		CreatorID: 42,
	}

	// Act
	output, err := usecase.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if !output.ChatRoom.IsMember(42) {
		t.Error("作成者がメンバーとして追加されていません")
	}
}

func TestCreateChatRoom_空文字の名前でルームを作成できない(t *testing.T) {
	// Arrange
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{userExists: true}
	usecase := createchatroom.NewUsecase(chatRoomRepo, userRepo)
	input := createchatroom.Input{
		Name:      "",
		CreatorID: 1,
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("エラーが発生するはずですが、発生しませんでした")
	}
	if err != chatroom.ErrEmptyName {
		t.Errorf("エラーが一致しません: got %v, want %v", err, chatroom.ErrEmptyName)
	}
}

func TestCreateChatRoom_存在しないユーザーIDではルームを作成できない(t *testing.T) {
	// Arrange
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{userExists: false}
	usecase := createchatroom.NewUsecase(chatRoomRepo, userRepo)
	input := createchatroom.Input{
		Name:      "テストルーム",
		CreatorID: 999,
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("エラーが発生するはずですが、発生しませんでした")
	}
	if err != createchatroom.ErrUserNotFound {
		t.Errorf("エラーが一致しません: got %v, want %v", err, createchatroom.ErrUserNotFound)
	}
}

func TestCreateChatRoom_ChatRoomCreatedイベントが発行される(t *testing.T) {
	// Arrange
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{userExists: true}
	usecase := createchatroom.NewUsecase(chatRoomRepo, userRepo)
	input := createchatroom.Input{
		Name:      "テストルーム",
		CreatorID: 1,
	}

	// Act
	output, err := usecase.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	events := output.ChatRoom.Events()
	var foundEvent *chatroom.ChatRoomCreatedEvent
	for _, e := range events {
		if evt, ok := e.(*chatroom.ChatRoomCreatedEvent); ok {
			foundEvent = evt
			break
		}
	}
	if foundEvent == nil {
		t.Fatal("ChatRoomCreatedイベントが発行されていません")
	}
	if foundEvent.Name != "テストルーム" {
		t.Errorf("イベントのNameが一致しません: got %s, want %s", foundEvent.Name, "テストルーム")
	}
}

func TestCreateChatRoom_MemberAddedイベントが発行される(t *testing.T) {
	// Arrange
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{userExists: true}
	usecase := createchatroom.NewUsecase(chatRoomRepo, userRepo)
	input := createchatroom.Input{
		Name:      "テストルーム",
		CreatorID: 42,
	}

	// Act
	output, err := usecase.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	events := output.ChatRoom.Events()
	var foundEvent *chatroom.MemberAddedEvent
	for _, e := range events {
		if evt, ok := e.(*chatroom.MemberAddedEvent); ok {
			foundEvent = evt
			break
		}
	}
	if foundEvent == nil {
		t.Fatal("MemberAddedイベントが発行されていません")
	}
	if foundEvent.UserID != 42 {
		t.Errorf("イベントのUserIDが一致しません: got %d, want %d", foundEvent.UserID, 42)
	}
}
