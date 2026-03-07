package model

type RoomConnection struct {
	Edges    []*RoomEdge `json:"edges"`
	PageInfo *PageInfo   `json:"pageInfo"`
	// TotalCount *int32    `json:"totalCount,omitempty"`
}
