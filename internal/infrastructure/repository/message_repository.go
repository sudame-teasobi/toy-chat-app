package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/db"
	"github.com/sudame/chat/internal/domain/message"
)

var _ message.Repository = (*MessageRepository)(nil)

type MessageRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewMessageRepository(database *sql.DB) *MessageRepository {
	return &MessageRepository{
		db:      database,
		queries: db.New(database),
	}
}

// FindById implements [message.Repository].
func (r *MessageRepository) FindById(ctx context.Context, id string) (*message.Message, error) {
	row, err := r.queries.GetMessage(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, message.ErrNotFound
		}
		return nil, err
	}

	m := message.Message{
		ID:           row.ID,
		ChatRoomID:   row.ChatRoomID,
		AuthorUserID: row.UserID,
		Body:         row.Body,
	}

	return &m, nil
}

// Save implements [message.Repository].
func (r *MessageRepository) Save(ctx context.Context, m *message.Message) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)

	for _, event := range m.Events {
		envelope, err := event.ToEnvelope()
		if err != nil {
			return err
		}

		switch envelope.Type {
		case message.MessagePostedEventType:
			_, err = qtx.PostMessage(ctx, db.PostMessageParams{
				UserID:     m.AuthorUserID,
				ChatRoomID: m.ChatRoomID,
				ID:         m.ID,
				Body:       m.Body,
			})
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown event type: %s", envelope.Type)
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

	m.ClearEvents()
	return tx.Commit()
}
