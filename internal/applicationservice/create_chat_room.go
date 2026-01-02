package applicationservice

import (
	"context"

	"github.com/sudame/chat/internal/domain/chatroom"
	"github.com/sudame/chat/internal/domain/user"
)

type CreateChatRoomInput struct {
	Name      string
	CreatorID string
}

type CreateChatRoomOutput struct {
	ChatRoom *chatroom.ChatRoom
}

type CreateChatRoomUsecase struct {
	chatRoomRepo chatroom.Repository
	userRepo     user.Repository
}

func NewCreateChatRoomUsecase(chatRoomRepo chatroom.Repository, userRepo user.Repository) *CreateChatRoomUsecase {
	return &CreateChatRoomUsecase{
		chatRoomRepo: chatRoomRepo,
		userRepo:     userRepo,
	}
}

func (u *CreateChatRoomUsecase) Execute(ctx context.Context, input CreateChatRoomInput) (*CreateChatRoomOutput, error) {
	// 1. ユーザー存在確認（別の集約なので外部リポジトリ）
	_, err := u.userRepo.FindByID(ctx, input.CreatorID)
	if err != nil {
		return nil, err
	}

	// 2. 集約を生成（ビジネスロジックは集約内）
	room, err := chatroom.NewChatRoom(input.Name, input.CreatorID)
	if err != nil {
		return nil, err // ErrEmptyName from domain
	}

	// 3. 集約を保存
	if err := u.chatRoomRepo.Save(ctx, room); err != nil {
		return nil, err
	}

	return &CreateChatRoomOutput{ChatRoom: room}, nil
}
