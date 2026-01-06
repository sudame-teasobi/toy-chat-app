package applicationservice_test

import (
	"context"
	"testing"

	"github.com/sudame/chat/internal/applicationservice"
	"github.com/sudame/chat/internal/domain/chatroom"
	"github.com/sudame/chat/internal/domain/user"
)

func TestCreateChatRoom_正常にチャットルームを作成できる(t *testing.T) {
	// Arrange
	existingUser := user.ReconstructUser("user-1", "テストユーザー")
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{user: existingUser}
	usecase := applicationservice.NewCreateChatRoomUsecase(chatRoomRepo, userRepo)
	input := applicationservice.CreateChatRoomInput{
		Name:      "テストルーム",
		CreatorID: "user-1",
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

func TestCreateChatRoom_作成時はメンバーが空_非同期で追加される(t *testing.T) {
	// Arrange
	existingUser := user.ReconstructUser("user-42", "テストユーザー")
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{user: existingUser}
	usecase := applicationservice.NewCreateChatRoomUsecase(chatRoomRepo, userRepo)
	input := applicationservice.CreateChatRoomInput{
		Name:      "テストルーム",
		CreatorID: "user-42",
	}

	// Act
	output, err := usecase.Execute(context.Background(), input)
	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	// メンバーは非同期で追加されるため、作成直後は空
	if len(output.ChatRoom.Members()) != 0 {
		t.Errorf("作成直後はメンバーが空であるべき: got %d members", len(output.ChatRoom.Members()))
	}
}

func TestCreateChatRoom_空文字の名前でルームを作成できない(t *testing.T) {
	// Arrange
	existingUser := user.ReconstructUser("user-1", "テストユーザー")
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{user: existingUser}
	usecase := applicationservice.NewCreateChatRoomUsecase(chatRoomRepo, userRepo)
	input := applicationservice.CreateChatRoomInput{
		Name:      "",
		CreatorID: "user-1",
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
	userRepo := &mockUserRepository{user: nil}
	usecase := applicationservice.NewCreateChatRoomUsecase(chatRoomRepo, userRepo)
	input := applicationservice.CreateChatRoomInput{
		Name:      "テストルーム",
		CreatorID: "user-999",
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("エラーが発生するはずですが、発生しませんでした")
	}
	if err != user.ErrNotFound {
		t.Errorf("エラーが一致しません: got %v, want %v", err, user.ErrNotFound)
	}
}

func TestCreateChatRoom_ChatRoomCreatedイベントが発行される(t *testing.T) {
	// Arrange
	existingUser := user.ReconstructUser("user-1", "テストユーザー")
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{user: existingUser}
	usecase := applicationservice.NewCreateChatRoomUsecase(chatRoomRepo, userRepo)
	input := applicationservice.CreateChatRoomInput{
		Name:      "テストルーム",
		CreatorID: "user-1",
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

func TestCreateChatRoom_ChatRoomCreatedEventにCreatorUserIDが含まれる(t *testing.T) {
	// Arrange
	existingUser := user.ReconstructUser("user-42", "テストユーザー")
	chatRoomRepo := &mockChatRoomRepository{}
	userRepo := &mockUserRepository{user: existingUser}
	usecase := applicationservice.NewCreateChatRoomUsecase(chatRoomRepo, userRepo)
	input := applicationservice.CreateChatRoomInput{
		Name:      "テストルーム",
		CreatorID: "user-42",
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
	if foundEvent.CreatorUserID != "user-42" {
		t.Errorf("イベントのCreatorUserIDが一致しません: got %s, want %s", foundEvent.CreatorUserID, "user-42")
	}
}
