package usecases_test

import (
	"context"

	"github.com/sudame/chat/internal/domain/chatroom"
	"github.com/sudame/chat/internal/domain/user"
)

// モックリポジトリ
type mockChatRoomRepository struct {
	room      *chatroom.ChatRoom
	savedRoom *chatroom.ChatRoom
}

func (m *mockChatRoomRepository) FindByID(ctx context.Context, id int64) (*chatroom.ChatRoom, error) {
	if m.room == nil {
		return nil, chatroom.ErrNotFound
	}

	return m.room, nil
}

func (m *mockChatRoomRepository) Save(ctx context.Context, room *chatroom.ChatRoom) error {
	m.savedRoom = room
	return nil
}

type mockUserRepository struct {
	user      *user.User
	savedUser *user.User
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int64) (*user.User, error) {
	if m.user == nil {
		return nil, user.ErrNotFound
	}

	return m.user, nil
}

func (m *mockUserRepository) Save(ctx context.Context, user *user.User) error {
	m.savedUser = user
	return nil
}
