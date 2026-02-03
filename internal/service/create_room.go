package service

import (
	"context"
	"fmt"

	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/internal/domain/user"
	"github.com/sudame/chat/internal/util"
)

type CreateRoomService struct {
	userRepo user.Repository
	roomRepo room.Repository
}

func NewCreateRoomService(userRepo user.Repository, roomRepo room.Repository) *CreateRoomService {
	return &CreateRoomService{
		userRepo: userRepo,
		roomRepo: roomRepo,
	}
}

func (s *CreateRoomService) Exec(ctx context.Context, roomName string, creatorUserId string) (*string, error) {
	_, err := s.userRepo.FindByID(ctx, creatorUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	r, err := room.NewRoom(roomName, creatorUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	err = s.roomRepo.Save(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("failed to save room: %w", err)
	}

	return util.ToPtr(r.ID()), nil
}
