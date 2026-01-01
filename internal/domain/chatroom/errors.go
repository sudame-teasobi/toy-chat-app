package chatroom

import "errors"

var (
	ErrEmptyName     = errors.New("チャットルーム名は空にできません")
	ErrAlreadyMember = errors.New("すでにメンバーです")
)
