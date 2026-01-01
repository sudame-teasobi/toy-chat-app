package addchatroommember

import (
	"context"

	"github.com/sudame/chat/internal/events"
	"github.com/sudame/chat/internal/models"
)

type Repository interface {
	UserExists(ctx context.Context, userID int64) (bool, error)
	ChatRoomExists(ctx context.Context, chatRoomID int64) (bool, error)
	IsMember(ctx context.Context, chatRoomID, userID int64) (bool, error)
	AddMember(ctx context.Context, chatRoomID, userID int64) (*models.ChatRoomMember, error)
	SaveEvent(ctx context.Context, event events.Event) error
}
