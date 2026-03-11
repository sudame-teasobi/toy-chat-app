package membership

import "context"

type Repository interface {
	Save(ctx context.Context, membership *Membership) error
	FindById(ctx context.Context, id string) (*Membership, error)
	FindByUserIDAndRoomID(ctx context.Context, userID string, roomID string) (*Membership, error)
}
