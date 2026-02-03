package service

import (
	"context"
	"fmt"

	"github.com/sudame/chat/internal/domain/user"
	"github.com/sudame/chat/internal/util"
)

type CreateUserService struct {
	userRepo user.Repository
}

func NewCreateUserService(userRepo user.Repository) *CreateUserService {
	return &CreateUserService{
		userRepo: userRepo,
	}
}

func (s *CreateUserService) Exec(ctx context.Context, name string) (*string, error) {
	u, err := user.NewUser(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	err = s.userRepo.Save(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return util.ToPtr(u.ID()), err
}
