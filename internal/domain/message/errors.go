package message

import "errors"

var (
	ErrNotFound  = errors.New("message not found")
	ErrForbidden = errors.New("permission missing")
)
