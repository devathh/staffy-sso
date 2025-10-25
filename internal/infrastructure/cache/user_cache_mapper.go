// Package cache implements mapper n' model for cache-storage
package cache

import (
	"encoding/json"

	domain "github.com/devathh/staffy-sso/internal/domain/user"
	"github.com/devathh/staffy-sso/pkg/consts"
)

type CacheMapper struct {
}

func (c *CacheMapper) ToModel(user *domain.User) (*UserModel, error) {
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

func (c *CacheMapper) ToDomain(data []byte) (*domain.User, error) {
	var result UserModel
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	email, err := domain.NewEmail(result.Email)
	if err != nil {
		return nil, err
	}

	return domain.FromPersistence(
		result.ID,
		email,
		result.Name,
		result.Surname,
		result.Password,
		result.IsRecruiter,
	), nil
}
