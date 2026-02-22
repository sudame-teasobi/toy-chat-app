package user

import (
	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/events"
)

type User struct {
	id     string
	name   string
	events []events.Event
}

func NewUser(name string) (*User, error) {
	id := "user:" + ulid.Make().String()

	if name == "" {
		return nil, ErrEmptyName
	}

	usr := &User{
		id:   id,
		name: name,
		events: []events.Event{
			&UserCreatedEvent{
				UserID: id,
				Name:   name,
			},
		},
	}

	return usr, nil
}

// ReconstructUser reconstructs a User from persistence.
func ReconstructUser(id string, name string) *User {
	return &User{
		id:   id,
		name: name,
	}
}

func (u *User) ID() string             { return u.id }
func (u *User) Name() string           { return u.name }
func (u *User) Events() []events.Event { return u.events }
