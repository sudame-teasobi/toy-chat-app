package room

import "context"

// Repository defines the interface for ChatRoom persistence.
type Repository interface {
	Save(ctx context.Context, chatRoom *Room) error
	FindByID(ctx context.Context, id string) (*Room, error)
}
