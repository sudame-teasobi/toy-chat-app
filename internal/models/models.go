package models

type ChatRoomMember struct {
	ID         int64
	ChatRoomID int64
	UserID     int64
}

type ChatRoom struct {
	ID   int64
	Name string
}
