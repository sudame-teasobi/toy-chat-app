package addchatroommember

import "errors"

var (
	ErrUserNotFound     = errors.New("ユーザーが存在しません")
	ErrChatRoomNotFound = errors.New("チャットルームが存在しません")
)
