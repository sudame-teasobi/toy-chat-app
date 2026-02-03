package repository

import (
	"context"
	"database/sql"

	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/db"
	"github.com/sudame/chat/internal/domain/room"
)

var _ room.Repository = (*ChatRoomRepository)(nil)

// ChatRoomRepository implements chatroom.Repository using TiDB.
type ChatRoomRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// NewChatRoomRepository creates a new ChatRoomRepository.
func NewChatRoomRepository(database *sql.DB) *ChatRoomRepository {
	return &ChatRoomRepository{
		db:      database,
		queries: db.New(database),
	}
}

// Save persists a chat room and its events to TiDB.
// It processes events to determine which operations to perform.
func (r *ChatRoomRepository) Save(ctx context.Context, cr *room.Room) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)

	// Process events to determine operations
	for _, event := range cr.Events() {
		envelope, err := event.ToEnvelope()
		if err != nil {
			return err
		}
		switch envelope.Type {
		case room.ChatRoomCreatedEventType:
			// Insert chat room only when ChatRoomCreatedEvent exists
			_, err = qtx.CreateChatRoom(ctx, db.CreateChatRoomParams{
				ID:   cr.ID(),
				Name: cr.Name(),
			})
			if err != nil {
				return err
			}
		}

		_, err = qtx.InsertEvent(ctx, db.InsertEventParams{
			ID:        ulid.Make().String(),
			EventType: envelope.Type,
			Payload:   envelope.Payload,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// FindByID retrieves a chat room by ID from TiDB.
func (r *ChatRoomRepository) FindByID(ctx context.Context, id string) (*room.Room, error) {
	row, err := r.queries.GetChatRoom(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, room.ErrNotFound
		}
		return nil, err
	}

	return room.ReconstructRoom(row.ID, row.Name), nil
}
