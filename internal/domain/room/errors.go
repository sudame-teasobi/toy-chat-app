package room

import "errors"

var (
	ErrEmptyName     = errors.New("チャットルーム名は空にできません")
	ErrAlreadyMember = errors.New("すでにメンバーです")
	ErrNotFound      = errors.New("チャットルームが存在しません")
	ErrNotAMember    = errors.New("チャットルームのメンバーではありません")
)
