// Package addchatroommember is a package.
package addchatroommember

import "context"

type Input struct {
	ChatRoomID int64
	UserID     int64
}

type Output struct {
	Member *ChatRoomMember
}

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Execute(ctx context.Context, input Input) (*Output, error) {
	exists, err := u.repo.UserExists(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	roomExists, err := u.repo.ChatRoomExists(ctx, input.ChatRoomID)
	if err != nil {
		return nil, err
	}
	if !roomExists {
		return nil, ErrChatRoomNotFound
	}

	isMember, err := u.repo.IsMember(ctx, input.ChatRoomID, input.UserID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, ErrAlreadyMember
	}

	member, err := u.repo.AddMember(ctx, input.ChatRoomID, input.UserID)
	if err != nil {
		return nil, err
	}

	if err := u.repo.SaveEvent(ctx, &MemberAddedEvent{
		ChatRoomID: input.ChatRoomID,
		UserID:     input.UserID,
	}); err != nil {
		return nil, err
	}

	return &Output{Member: member}, nil
}
