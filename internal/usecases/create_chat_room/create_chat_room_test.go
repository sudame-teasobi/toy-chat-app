package createchatroom_test

import (
	"context"
	"testing"

	createchatroom "github.com/sudame/chat/internal/usecases/create_chat_room"
)

// モックリポジトリ
type mockChatRoomRepository struct {
	createdRoom   *createchatroom.ChatRoom
	createdMember *createchatroom.ChatRoomMember
	userExists    bool
	savedEvents   []createchatroom.Event
}

func (m *mockChatRoomRepository) CreateChatRoom(ctx context.Context, name string) (*createchatroom.ChatRoom, error) {
	m.createdRoom = &createchatroom.ChatRoom{
		ID:   1,
		Name: name,
	}
	return m.createdRoom, nil
}

func (m *mockChatRoomRepository) AddMember(ctx context.Context, chatRoomID, userID int64) (*createchatroom.ChatRoomMember, error) {
	m.createdMember = &createchatroom.ChatRoomMember{
		ID:         1,
		ChatRoomID: chatRoomID,
		UserID:     userID,
	}
	return m.createdMember, nil
}

func (m *mockChatRoomRepository) UserExists(ctx context.Context, userID int64) (bool, error) {
	return m.userExists, nil
}

func (m *mockChatRoomRepository) SaveEvent(ctx context.Context, event createchatroom.Event) error {
	m.savedEvents = append(m.savedEvents, event)
	return nil
}

func TestCreateChatRoom_正常にチャットルームを作成できる(t *testing.T) {
	// Arrange
	repo := &mockChatRoomRepository{userExists: true}
	usecase := createchatroom.NewUsecase(repo)
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
	if output.ChatRoom.Name != "テストルーム" {
		t.Errorf("ルーム名が一致しません: got %s, want %s", output.ChatRoom.Name, "テストルーム")
	}
	if output.ChatRoom.ID == 0 {
		t.Error("ルームIDが設定されていません")
	}
}

func TestCreateChatRoom_作成者が自動的にメンバーとして追加される(t *testing.T) {
	// Arrange
	repo := &mockChatRoomRepository{userExists: true}
	usecase := createchatroom.NewUsecase(repo)
	input := createchatroom.Input{
		Name:      "テストルーム",
		CreatorID: 42,
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if repo.createdMember == nil {
		t.Fatal("メンバーが作成されていません")
	}
	if repo.createdMember.UserID != 42 {
		t.Errorf("メンバーのUserIDが一致しません: got %d, want %d", repo.createdMember.UserID, 42)
	}
	if repo.createdMember.ChatRoomID != repo.createdRoom.ID {
		t.Errorf("メンバーのChatRoomIDが一致しません: got %d, want %d", repo.createdMember.ChatRoomID, repo.createdRoom.ID)
	}
}

func TestCreateChatRoom_空文字の名前でルームを作成できない(t *testing.T) {
	// Arrange
	repo := &mockChatRoomRepository{userExists: true}
	usecase := createchatroom.NewUsecase(repo)
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
	if err != createchatroom.ErrEmptyName {
		t.Errorf("エラーが一致しません: got %v, want %v", err, createchatroom.ErrEmptyName)
	}
}

func TestCreateChatRoom_存在しないユーザーIDではルームを作成できない(t *testing.T) {
	// Arrange
	repo := &mockChatRoomRepository{userExists: false}
	usecase := createchatroom.NewUsecase(repo)
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
	repo := &mockChatRoomRepository{userExists: true}
	usecase := createchatroom.NewUsecase(repo)
	input := createchatroom.Input{
		Name:      "テストルーム",
		CreatorID: 1,
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	// ChatRoomCreatedイベントが発行されているか確認
	var foundEvent *createchatroom.ChatRoomCreatedEvent
	for _, e := range repo.savedEvents {
		if evt, ok := e.(*createchatroom.ChatRoomCreatedEvent); ok {
			foundEvent = evt
			break
		}
	}
	if foundEvent == nil {
		t.Fatal("ChatRoomCreatedイベントが発行されていません")
	}
	if foundEvent.ChatRoomID != repo.createdRoom.ID {
		t.Errorf("イベントのChatRoomIDが一致しません: got %d, want %d", foundEvent.ChatRoomID, repo.createdRoom.ID)
	}
	if foundEvent.Name != "テストルーム" {
		t.Errorf("イベントのNameが一致しません: got %s, want %s", foundEvent.Name, "テストルーム")
	}
}

func TestCreateChatRoom_MemberAddedイベントが発行される(t *testing.T) {
	// Arrange
	repo := &mockChatRoomRepository{userExists: true}
	usecase := createchatroom.NewUsecase(repo)
	input := createchatroom.Input{
		Name:      "テストルーム",
		CreatorID: 42,
	}

	// Act
	_, err := usecase.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	// MemberAddedイベントが発行されているか確認
	var foundEvent *createchatroom.MemberAddedEvent
	for _, e := range repo.savedEvents {
		if evt, ok := e.(*createchatroom.MemberAddedEvent); ok {
			foundEvent = evt
			break
		}
	}
	if foundEvent == nil {
		t.Fatal("MemberAddedイベントが発行されていません")
	}
	if foundEvent.ChatRoomID != repo.createdRoom.ID {
		t.Errorf("イベントのChatRoomIDが一致しません: got %d, want %d", foundEvent.ChatRoomID, repo.createdRoom.ID)
	}
	if foundEvent.UserID != 42 {
		t.Errorf("イベントのUserIDが一致しません: got %d, want %d", foundEvent.UserID, 42)
	}
}
