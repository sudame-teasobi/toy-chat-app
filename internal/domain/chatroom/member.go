package chatroom

// Member represents a chat room member.
type Member struct {
	id     int64
	userID int64
}

// NewMember creates a new member.
func NewMember(userID int64) Member {
	return Member{userID: userID}
}

// ReconstructMember reconstructs a member from persistence.
func ReconstructMember(id, userID int64) Member {
	return Member{id: id, userID: userID}
}

func (m Member) ID() int64     { return m.id }
func (m Member) UserID() int64 { return m.userID }
