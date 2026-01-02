package chatroom

import "context"

// Repository defines the interface for ChatRoom persistence.
type Repository interface {
	Save(ctx context.Context, chatRoom *ChatRoom) error
	FindByID(ctx context.Context, id string) (*ChatRoom, error)
}
