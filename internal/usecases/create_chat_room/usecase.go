package createchatroom

import (
	"context"

	"github.com/sudame/chat/internal/domain/chatroom"
)

type Input struct {
	Name      string
	CreatorID int64
}

type Output struct {
	ChatRoom *chatroom.ChatRoom
}

type Usecase struct {
	chatRoomRepo chatroom.Repository
	userRepo     UserRepository
}

func NewUsecase(chatRoomRepo chatroom.Repository, userRepo UserRepository) *Usecase {
	return &Usecase{
		chatRoomRepo: chatRoomRepo,
		userRepo:     userRepo,
	}
}

func (u *Usecase) Execute(ctx context.Context, input Input) (*Output, error) {
	// 1. ユーザー存在確認（別の集約なので外部リポジトリ）
	exists, err := u.userRepo.UserExists(ctx, input.CreatorID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	// 2. 集約を生成（ビジネスロジックは集約内）
	room, err := chatroom.NewChatRoom(0, input.Name, input.CreatorID)
	if err != nil {
		return nil, err // ErrEmptyName from domain
	}

	// 3. 集約を保存
	if err := u.chatRoomRepo.Save(ctx, room); err != nil {
		return nil, err
	}

	return &Output{ChatRoom: room}, nil
}
