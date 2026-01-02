package user

import "context"

// Repository defines the interface for ChatRoom persistence.
type Repository interface {
	Save(ctx context.Context, chatRoom *User) error
	FindByID(ctx context.Context, id int64) (*User, error)
}
