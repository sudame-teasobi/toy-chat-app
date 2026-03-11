package membership

type CheckMembershipExistenceRequest struct {
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
}

type CheckMembershipExistenceResponse struct {
	Existence bool `json:"existence"`
}

type Query interface {
	CheckMembershipExistence(req CheckMembershipExistenceRequest) (CheckMembershipExistenceResponse, error)
}
