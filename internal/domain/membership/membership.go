package membership

import "github.com/oklog/ulid/v2"

type Membership struct {
	Id         string `json:"id"`
	ChatRoomId string `json:"chatRoomId"`
	UserId     string `json:"userId"`
}

func NewMembership(chatRoomId string, userId string) (*Membership, error) {
	id := "membership-" + ulid.Make().String()

	return &Membership{
		Id:         id,
		ChatRoomId: chatRoomId,
		UserId:     userId,
	}, nil
}
