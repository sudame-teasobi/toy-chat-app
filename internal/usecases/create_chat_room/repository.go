package createchatroom

import "context"

type ChatRoom struct {
	ID   int64
	Name string
}

type ChatRoomMember struct {
	ID         int64
	ChatRoomID int64
	UserID     int64
}

type Repository interface {
	CreateChatRoom(ctx context.Context, name string) (*ChatRoom, error)
	AddMember(ctx context.Context, chatRoomID, userID int64) (*ChatRoomMember, error)
	UserExists(ctx context.Context, userID int64) (bool, error)
	SaveEvent(ctx context.Context, event Event) error
}
