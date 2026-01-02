package chatroom

// Member represents a chat room member.
type Member struct {
	id     string
	userID string
}

// NewMember creates a new member.
func NewMember(userID string) Member {
	return Member{userID: userID}
}

// ReconstructMember reconstructs a member from persistence.
func ReconstructMember(id, userID string) Member {
	return Member{id: id, userID: userID}
}

func (m Member) ID() string     { return m.id }
func (m Member) UserID() string { return m.userID }
