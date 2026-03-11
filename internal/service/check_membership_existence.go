package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/sudame/chat/internal/domain/membership"
)

type CheckMembershipExistenceService struct {
	membershipRepo membership.Repository
}

func NewCheckMembershipExistenceService(membershipRepo membership.Repository) *CheckMembershipExistenceService {
	return &CheckMembershipExistenceService{
		membershipRepo: membershipRepo,
	}
}

type CheckMembershipExistenceServiceResult struct {
	Existence bool
}

func (s *CheckMembershipExistenceService) Exec(ctx context.Context, userID string, roomID string) (CheckMembershipExistenceServiceResult, error) {
	var zero CheckMembershipExistenceServiceResult
	var existence bool

	_, err := s.membershipRepo.FindByUserIDAndRoomID(ctx, userID, roomID)

	if err == nil {
		existence = true
	} else if errors.Is(err, membership.ErrNotFound) {
		existence = false
	} else {
		return zero, fmt.Errorf("failed to find membership: %w", err)
	}

	result := CheckMembershipExistenceServiceResult{Existence: existence}
	return result, nil
}
