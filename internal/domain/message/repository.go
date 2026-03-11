package message

import "context"

type Repository interface {
	Save(ctx context.Context, message *Message) error
	FindById(ctx context.Context, id string) (*Message, error)
}
