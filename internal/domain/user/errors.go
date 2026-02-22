package user

import "errors"

var (
	ErrNotFound  = errors.New("user not found")
	ErrEmptyName = errors.New("user name is empty")
)
