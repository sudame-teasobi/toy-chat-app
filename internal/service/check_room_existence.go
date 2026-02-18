package service

import (
	"context"
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

func (s *CheckRoomExistenceService) Exec(ctx context.Context, roomId string) CheckRoomExistenceServiceResult {
	_, err := s.roomRepo.FindByID(ctx, roomId)
	result := CheckRoomExistenceServiceResult{Existence: err == nil}
	slog.DebugContext(ctx, "CheckRoomExistenceService executed", "room_id", roomId, "result", result)
	return result
}
