package user

import "context"

// Repository defines the interface for User persistence.
type Repository interface {
	Save(ctx context.Context, chatRoom *User) error
	FindByID(ctx context.Context, id string) (*User, error)
}
