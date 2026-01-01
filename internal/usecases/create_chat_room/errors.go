package createchatroom

import "errors"

var (
	ErrEmptyName    = errors.New("チャットルーム名は空にできません")
	ErrUserNotFound = errors.New("ユーザーが存在しません")
)
