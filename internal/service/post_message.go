package service

import (
	"context"
	"fmt"

	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/internal/domain/message"
)

type PostMessageService struct {
	membershipQuery membership.Query
	messageRepo     message.Repository
}

func NewPostMessageService(_ context.Context, membershipQuery membership.Query, messageRepo message.Repository) *PostMessageService {
	return &PostMessageService{
		membershipQuery: membershipQuery,
		messageRepo:     messageRepo,
	}
}

func (s *PostMessageService) Exec(ctx context.Context, authorUserID string, roomID string, body string) (*string, error) {
	// メンバーシップの存在確認
	membershipExistence, err := s.membershipQuery.CheckMembershipExistence(membership.CheckMembershipExistenceRequest{
		RoomID: roomID,
		UserID: authorUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call membership service: %w", err)
	}
	if !membershipExistence.Existence {
		return nil, fmt.Errorf("failed to find membership (userID = %s, roomID = %s)", authorUserID, roomID)
	}

	m, err := message.PostMessage(roomID, authorUserID, body)
	if err != nil {
		return nil, fmt.Errorf("failed to post message: %w", err)
	}

	err = s.messageRepo.Save(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	return &m.ID, nil
}
