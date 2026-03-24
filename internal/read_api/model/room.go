package model

type Room struct {
	ID      string                `json:"id"`
	Name    string                `json:"name"`
	Members *RoomMemberConnection `json:"members"`
}

func (Room) IsNode()         {}
func (r Room) GetID() string { return r.ID }

type RoomConnection struct {
	Edges    []*RoomEdge `json:"edges"`
	PageInfo *PageInfo   `json:"pageInfo"`
	// TotalCount *int32    `json:"totalCount,omitempty"`
}
