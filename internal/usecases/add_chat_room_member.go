package usecases

import (
	"context"

	"github.com/sudame/chat/internal/domain/chatroom"
	"github.com/sudame/chat/internal/domain/user"
)

type AddChatRoomMemberInput struct {
	ChatRoomID int64
	UserID     int64
}

type AddChatRoomMemberOutput struct {
	ChatRoom *chatroom.ChatRoom
}

type AddChatRoomMemberUsecase struct {
	chatRoomRepo chatroom.Repository
	userRepo     user.Repository
}

func NewAddChatRoomMemberUsecase(chatRoomRepo chatroom.Repository, userRepo user.Repository) *AddChatRoomMemberUsecase {
	return &AddChatRoomMemberUsecase{
		chatRoomRepo: chatRoomRepo,
		userRepo:     userRepo,
	}
}

func (u *AddChatRoomMemberUsecase) Execute(ctx context.Context, input AddChatRoomMemberInput) (*AddChatRoomMemberOutput, error) {
	// 1. ユーザー存在確認
	usr, err := u.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if usr == nil {
		return nil, user.ErrNotFound
	}

	// 2. 集約をロード
	room, err := u.chatRoomRepo.FindByID(ctx, input.ChatRoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, chatroom.ErrNotFound
	}

	// 3. 集約にビジネスロジックを実行
	if err := room.AddMember(input.UserID); err != nil {
		return nil, err // ErrAlreadyMember from domain
	}

	// 4. 集約を保存
	if err := u.chatRoomRepo.Save(ctx, room); err != nil {
		return nil, err
	}

	return &AddChatRoomMemberOutput{ChatRoom: room}, nil
}
