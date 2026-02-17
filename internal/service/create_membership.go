package service

import (
	"context"
	"fmt"

	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/internal/domain/user"
)

type CreateMembershipService struct {
	userRepo       user.Repository
	roomQuery      room.Query
	membershipRepo membership.Repository
}

func NewCreateMembershipService(userRepo user.Repository, roomQuery room.Query, membershipRepo membership.Repository) *CreateMembershipService {
	return &CreateMembershipService{
		userRepo:       userRepo,
		roomQuery:      roomQuery,
		membershipRepo: membershipRepo,
	}
}

func (s *CreateMembershipService) Exec(ctx context.Context, userID string, roomID string) error {
	// ユーザーの存在確認
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// ルームの存在確認
	roomExistence, err := s.roomQuery.CheckRoomExistence(room.CheckRoomExistenceRequest{RoomID: roomID})
	if err != nil {
		return fmt.Errorf("failed to call room service: %w", err)
	}
	if !roomExistence.Existence {
		return fmt.Errorf("failed to find room %s", roomID)
	}

	m, err := membership.CreateMembership(roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to create memberhsip: %w", err)
	}

	err = s.membershipRepo.Save(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to save membership: %w", err)
	}

	return nil
}
