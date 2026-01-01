package createchatroom

import "context"

type Input struct {
	Name      string
	CreatorID int64
}

type Output struct {
	ChatRoom *ChatRoom
}

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Execute(ctx context.Context, input Input) (*Output, error) {
	if input.Name == "" {
		return nil, ErrEmptyName
	}

	exists, err := u.repo.UserExists(ctx, input.CreatorID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	room, err := u.repo.CreateChatRoom(ctx, input.Name)
	if err != nil {
		return nil, err
	}

	if err := u.repo.SaveEvent(ctx, &ChatRoomCreatedEvent{
		ChatRoomID: room.ID,
		Name:       room.Name,
	}); err != nil {
		return nil, err
	}

	_, err = u.repo.AddMember(ctx, room.ID, input.CreatorID)
	if err != nil {
		return nil, err
	}

	if err := u.repo.SaveEvent(ctx, &MemberAddedEvent{
		ChatRoomID: room.ID,
		UserID:     input.CreatorID,
	}); err != nil {
		return nil, err
	}

	return &Output{ChatRoom: room}, nil
}
