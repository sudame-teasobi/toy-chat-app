package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/db"
	"github.com/sudame/chat/internal/domain/membership"
)

var _ membership.Repository = (*MembershipRepository)(nil)

type MembershipRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewMembershipRepository(database *sql.DB) *MembershipRepository {
	return &MembershipRepository{
		db:      database,
		queries: db.New(database),
	}
}

// FindById implements [membership.Repository].
func (r *MembershipRepository) FindById(ctx context.Context, id string) (*membership.Membership, error) {
	row, err := r.queries.GetMember(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, membership.ErrNotFound
		}
		return nil, err
	}

	ms := membership.Membership{
		Id:         row.ID,
		ChatRoomId: row.ChatRoomID,
		UserId:     row.UserID,
	}

	return &ms, nil
}

// Save implements [membership.Repository].
func (r *MembershipRepository) Save(ctx context.Context, m *membership.Membership) error {
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
		case membership.MembershipCreatedEventType:
			_, err = qtx.CreateMember(ctx, db.CreateMemberParams{
				UserID:     m.UserId,
				ChatRoomID: m.ChatRoomId,
				ID:         m.Id,
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

	return tx.Commit()
}
