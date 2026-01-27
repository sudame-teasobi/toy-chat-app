package repository

import (
	"context"
	"database/sql"

	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/db"
	"github.com/sudame/chat/internal/domain/user"
)

var _ user.Repository = (*UserRepository)(nil)

// UserRepository implements user.Repository using TiDB.
type UserRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(database *sql.DB) *UserRepository {
	return &UserRepository{
		db:      database,
		queries: db.New(database),
	}
}

// Save persists a user and its events to TiDB.
func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)

	// Insert user
	_, err = qtx.CreateUser(ctx, db.CreateUserParams{
		ID:   u.ID(),
		Name: u.Name(),
	})
	if err != nil {
		return err
	}

	// Insert events
	for _, event := range u.Events() {
		eventEnvelope, err := event.ToEnvelope()
		if err != nil {
			return err
		}

		_, err = qtx.InsertEvent(ctx, db.InsertEventParams{
			ID:        ulid.Make().String(),
			EventType: eventEnvelope.Type,
			Payload:   eventEnvelope.Payload,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// FindByID retrieves a user by ID from TiDB.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	row, err := r.queries.GetUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrNotFound
		}
		return nil, err
	}

	return user.ReconstructUser(row.ID, row.Name), nil
}
