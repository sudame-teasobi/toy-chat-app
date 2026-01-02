package user

import "github.com/sudame/chat/internal/events"

// User is the aggregate root for user domain.
type User struct {
	id     int64
	name   string
	events []events.Event
}

func NewUser(id int64, name string) (*User, error) {
	if name == "" {
		return nil, ErrEmptyName
	}

	usr := &User{
		id:     id,
		name:   name,
		events: make([]events.Event, 0),
	}

	usr.events = append(usr.events, &UserCreatedEvent{
		UserID: usr.id,
		Name:   usr.name,
	})

	return usr, nil
}

// ReconstructChatRoom reconstructs a ChatRoom from persistence.
func ReconstructUser(id int64, name string) *User {
	return &User{
		id:   id,
		name: name,
	}
}

func (u *User) ID() int64              { return u.id }
func (u *User) Name() string           { return u.name }
func (u *User) Events() []events.Event { return u.events }
