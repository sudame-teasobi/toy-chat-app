package query

import (
	"fmt"

	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/pkg/httpclient"
)

type RoomQuery struct {
	client  *httpclient.HTTPClient
	baseURL string
}

func NewRoomQuery(client *httpclient.HTTPClient, baseURL string) *RoomQuery {
	return &RoomQuery{client: client, baseURL: baseURL}
}

const checkRoomExistencePath = "/check-room-existence"

// CheckRoomExistence implements [room.Query].
func (r *RoomQuery) CheckRoomExistence(req room.CheckRoomExistenceRequest) (room.CheckRoomExistenceResponse, error) {
	var zero room.CheckRoomExistenceResponse
	res, err := httpclient.Post[room.CheckRoomExistenceRequest, room.CheckRoomExistenceResponse](r.client, r.baseURL+checkRoomExistencePath, req)
	if err != nil {
		return zero, fmt.Errorf("failed to post: %w", err)
	}

	return res, nil
}

var _ room.Query = (*RoomQuery)(nil)
