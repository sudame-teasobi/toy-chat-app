package room

import "errors"

var (
	ErrEmptyName = errors.New("room name is empty")
	ErrNotFound  = errors.New("room not found")
)
