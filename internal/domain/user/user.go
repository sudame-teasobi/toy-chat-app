package user

// User is the aggregate root for user domain.
type User struct {
	id   int64
	name string
}

// ReconstructChatRoom reconstructs a ChatRoom from persistence.
func ReconstructUser(id int64, name string) *User {
	return &User{
		id:   id,
		name: name,
	}
}

func (u *User) ID() int64    { return u.id }
func (u *User) Name() string { return u.name }
