package user

import "errors"

var (
	ErrNotFound  = errors.New("ユーザーが存在しません")
	ErrEmptyName = errors.New("ユーザー名が空です")
)
