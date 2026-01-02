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

func TestCreateChatRoom_作成者が自動的にメンバーとして追加される(t *testing.T) {
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
	if !output.ChatRoom.IsMember("user-42") {
		t.Error("作成者がメンバーとして追加されていません")
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

func TestCreateChatRoom_MemberAddedイベントが発行される(t *testing.T) {
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
	if foundEvent.UserID != "user-42" {
		t.Errorf("イベントのUserIDが一致しません: got %s, want %s", foundEvent.UserID, "user-42")
	}
}
