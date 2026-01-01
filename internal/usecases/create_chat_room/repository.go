package createchatroom

import (
	"context"

	"github.com/sudame/chat/internal/events"
	"github.com/sudame/chat/internal/models"
)

// Repository defines the persistence operations required by this usecase.
type Repository interface {
	CreateChatRoom(ctx context.Context, name string) (*models.ChatRoom, error)
	AddMember(ctx context.Context, chatRoomID, userID int64) (*models.ChatRoomMember, error)
	UserExists(ctx context.Context, userID int64) (bool, error)
	SaveEvent(ctx context.Context, event events.Event) error
}
