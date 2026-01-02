package applicationservice_test

import (
	"context"
	"testing"

	"github.com/sudame/chat/internal/applicationservice"
	"github.com/sudame/chat/internal/domain/chatroom"
	"github.com/sudame/chat/internal/domain/user"
)

func TestAddChatRoomMember_正常にメンバーを追加できる(t *testing.T) {
	// Arrange
	existingRoom := chatroom.ReconstructChatRoom("room-1", "テストルーム", []chatroom.Member{
		chatroom.ReconstructMember("member-1", "user-100"), // 既存メンバー
	})
	existingUser := user.ReconstructUser("user-42", "テストユーザー")
	chatRoomRepo := &mockChatRoomRepository{room: existingRoom}
	userRepo := &mockUserRepository{user: existingUser}
	usecase := applicationservice.NewAddChatRoomMemberUsecase(chatRoomRepo, userRepo)
	input := applicationservice.AddChatRoomMemberInput{
		ChatRoomID: "room-1",
		UserID:     "user-42",
	}

	// Act
	output, err := usecase.Execute(context.Background(), input)
	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if !output.ChatRoom.IsMember("user-42") {
		t.Error("メンバーが追加されていません")
	}
}

func TestAddChatRoomMember_MemberAddedイベントが発行される(t *testing.T) {
	// Arrange
	existingRoom := chatroom.ReconstructChatRoom("room-1", "テストルーム", []chatroom.Member{
		chatroom.ReconstructMember("member-1", "user-100"),
	})
	chatRoomRepo := &mockChatRoomRepository{room: existingRoom}
	existingUser := user.ReconstructUser("user-42", "テストユーザー")
	userRepo := &mockUserRepository{user: existingUser}
	usecase := applicationservice.NewAddChatRoomMemberUsecase(chatRoomRepo, userRepo)
	input := applicationservice.AddChatRoomMemberInput{
		ChatRoomID: "room-1",
		UserID:     "user-42",
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
	if foundEvent.ChatRoomID != "room-1" {
		t.Errorf("イベントのChatRoomIDが一致しません: got %s, want %s", foundEvent.ChatRoomID, "room-1")
	}
	if foundEvent.UserID != "user-42" {
		t.Errorf("イベントのUserIDが一致しません: got %s, want %s", foundEvent.UserID, "user-42")
	}
}

func TestAddChatRoomMember_存在しないユーザーを追加できない(t *testing.T) {
	// Arrange
	existingRoom := chatroom.ReconstructChatRoom("room-1", "テストルーム", []chatroom.Member{})
	chatRoomRepo := &mockChatRoomRepository{room: existingRoom}
	userRepo := &mockUserRepository{user: nil}
	usecase := applicationservice.NewAddChatRoomMemberUsecase(chatRoomRepo, userRepo)
	input := applicationservice.AddChatRoomMemberInput{
		ChatRoomID: "room-1",
		UserID:     "user-999",
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

func TestAddChatRoomMember_存在しないチャットルームに追加できない(t *testing.T) {
	// Arrange
	chatRoomRepo := &mockChatRoomRepository{room: nil} // ルームが存在しない
	existingUser := user.ReconstructUser("user-1", "テストユーザー")
	userRepo := &mockUserRepository{user: existingUser}
	usecase := applicationservice.NewAddChatRoomMemberUsecase(chatRoomRepo, userRepo)
	input := applicationservice.AddChatRoomMemberInput{
		ChatRoomID: "room-999",
		UserID:     "user-1",
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("エラーが発生するはずですが、発生しませんでした")
	}
	if err != chatroom.ErrNotFound {
		t.Errorf("エラーが一致しません: got %v, want %v", err, chatroom.ErrNotFound)
	}
}

func TestAddChatRoomMember_すでにメンバーの場合は追加できない(t *testing.T) {
	// Arrange
	existingRoom := chatroom.ReconstructChatRoom("room-1", "テストルーム", []chatroom.Member{
		chatroom.ReconstructMember("member-1", "user-42"), // 既にメンバー
	})
	existingUser := user.ReconstructUser("user-42", "テストユーザー")
	chatRoomRepo := &mockChatRoomRepository{room: existingRoom}
	userRepo := &mockUserRepository{user: existingUser}
	usecase := applicationservice.NewAddChatRoomMemberUsecase(chatRoomRepo, userRepo)
	input := applicationservice.AddChatRoomMemberInput{
		ChatRoomID: "room-1",
		UserID:     "user-42",
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("エラーが発生するはずですが、発生しませんでした")
	}
	if err != chatroom.ErrAlreadyMember {
		t.Errorf("エラーが一致しません: got %v, want %v", err, chatroom.ErrAlreadyMember)
	}
}
