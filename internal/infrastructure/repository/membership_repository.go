package repository

import (
	"context"
	"database/sql"

	"github.com/sudame/chat/internal/db"
	"github.com/sudame/chat/internal/domain/membership"
)

var _ membership.Repository = (*MembershipRepository)(nil)

type MembershipRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// FindById implements [membership.Repository].
func (r *MembershipRepository) FindById(ctx context.Context, id string) (*membership.Membership, error) {
	row, err := r.queries.GetMember(ctx, id)
	if err != nil {
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
		}
	}

	return tx.Commit()
}
