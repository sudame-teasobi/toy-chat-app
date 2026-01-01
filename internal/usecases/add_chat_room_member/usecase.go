package addchatroommember

import (
	"context"

	"github.com/sudame/chat/internal/domain/chatroom"
)

type Input struct {
	ChatRoomID int64
	UserID     int64
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
	// 1. ユーザー存在確認
	exists, err := u.userRepo.UserExists(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	// 2. 集約をロード
	room, err := u.chatRoomRepo.FindByID(ctx, input.ChatRoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrChatRoomNotFound
	}

	// 3. 集約にビジネスロジックを実行
	if err := room.AddMember(input.UserID); err != nil {
		return nil, err // ErrAlreadyMember from domain
	}

	// 4. 集約を保存
	if err := u.chatRoomRepo.Save(ctx, room); err != nil {
		return nil, err
	}

	return &Output{ChatRoom: room}, nil
}
