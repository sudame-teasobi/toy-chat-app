package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/sudame/chat/internal/domain/room"
)

type CheckRoomExistenceService struct {
	roomRepo room.Repository
}

func NewCheckRoomExistenceService(roomRepo room.Repository) *CheckRoomExistenceService {
	return &CheckRoomExistenceService{
		roomRepo: roomRepo,
	}
}

type CheckRoomExistenceServiceResult struct {
	Existence bool
}

func (s *CheckRoomExistenceService) Exec(ctx context.Context, roomId string) (CheckRoomExistenceServiceResult, error) {
	var zero CheckRoomExistenceServiceResult
	var existence bool

	_, err := s.roomRepo.FindByID(ctx, roomId)

	if err == nil {
		existence = true
	} else if errors.Is(err, room.ErrNotFound) {
		existence = false
	} else {
		return zero, fmt.Errorf("failed to find room: %w", err)
	}

	result := CheckRoomExistenceServiceResult{Existence: existence}
	slog.DebugContext(ctx, "CheckRoomExistenceService executed", "room_id", roomId, "result", result)
	return result, nil
}
