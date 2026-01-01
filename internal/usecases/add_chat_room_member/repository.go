package addchatroommember

import "context"

type ChatRoomMember struct {
	ID         int64
	ChatRoomID int64
	UserID     int64
}

type Event interface {
	EventType() string
}

type Repository interface {
	UserExists(ctx context.Context, userID int64) (bool, error)
	ChatRoomExists(ctx context.Context, chatRoomID int64) (bool, error)
	IsMember(ctx context.Context, chatRoomID, userID int64) (bool, error)
	AddMember(ctx context.Context, chatRoomID, userID int64) (*ChatRoomMember, error)
	SaveEvent(ctx context.Context, event Event) error
}
