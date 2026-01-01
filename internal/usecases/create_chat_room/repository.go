package createchatroom

import "context"

// UserRepository handles user queries.
// TODO: User集約ができたらdomain/user/repository.goに移動
type UserRepository interface {
	UserExists(ctx context.Context, userID int64) (bool, error)
}
