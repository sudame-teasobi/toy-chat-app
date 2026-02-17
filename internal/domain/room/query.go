package room

type CheckRoomExistenceRequest struct {
	RoomID string `json:"room_id"`
}

type CheckRoomExistenceResponse struct {
	Existence bool `json:"existence"`
}

type Query interface {
	CheckRoomExistence(req CheckRoomExistenceRequest) (CheckRoomExistenceResponse, error)
}
