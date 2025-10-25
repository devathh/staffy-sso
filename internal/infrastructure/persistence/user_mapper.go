// Package persistence implements mapper n' model for db
package persistence

import (
	domain "github.com/devathh/staffy-sso/internal/domain/user"
	"github.com/devathh/staffy-sso/pkg/consts"
)

type UserMapper struct {
}

func (u *UserMapper) ToModel(user *domain.User) (*UserModel, error) {
	if user == nil {
		return nil, consts.ErrEmptyUser
	}

	return &UserModel{
		ID:          user.ID(),
		Email:       user.Email(),
		Name:        user.Name(),
		Surname:     user.Surname(),
		IsRecruiter: user.IsRecruiter(),
		Password:    user.Password(),
	}, nil
}

func (u *UserMapper) ToDomain(user *UserModel) (*domain.User, error) {
	if user == nil {
		return nil, consts.ErrEmptyUser
	}

	email, err := domain.NewEmail(user.Email)
	if err != nil {
		return nil, err
	}

	return domain.FromPersistence(
		user.ID, email, user.Name,
		user.Surname, user.Password, user.IsRecruiter,
	), nil
}
