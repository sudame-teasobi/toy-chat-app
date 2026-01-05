package applicationservice

import (
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/sudame/chat/internal/domain/user"
)

type CreateUserInput struct {
	Name string
}

type CreateUserOutput struct {
	User *user.User
}

type CreateUserUsecase struct {
	userRepo user.Repository
}

func NewCreateUserUsecase(userRepo user.Repository) *CreateUserUsecase {
	return &CreateUserUsecase{
		userRepo: userRepo,
	}
}

func (u *CreateUserUsecase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	// 1. IDを生成
	id := ulid.Make().String()

	// 2. ユーザー集約を生成
	usr, err := user.NewUser(id, input.Name)
	if err != nil {
		return nil, err
	}

	// 3. ユーザーを保存
	if err := u.userRepo.Save(ctx, usr); err != nil {
		return nil, err
	}

	return &CreateUserOutput{User: usr}, nil
}
