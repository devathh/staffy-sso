// Package redis implements structure which allows you to manage the cache
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	domain "github.com/devathh/staffy-sso/internal/domain/user"
	"github.com/devathh/staffy-sso/internal/infrastructure/cache"
	"github.com/devathh/staffy-sso/internal/infrastructure/config"
	"github.com/devathh/staffy-sso/pkg/consts"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type UserCache struct {
	client *redis.Client
	mapper *cache.CacheMapper
	cfg    *config.Config
}

func (u *UserCache) SetByEmail(ctx context.Context, user *domain.User) error {
	return u.set(ctx, user, u.userKey(user.Email()))
}

func (u *UserCache) SetByID(ctx context.Context, user *domain.User) error {
	return u.set(ctx, user, u.userKey(user.ID()))
}

func (u *UserCache) set(ctx context.Context, user *domain.User, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	userModel, err := u.mapper.ToModel(user)
	if err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	data, err := json.Marshal(userModel)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	status := u.client.Set(ctx, u.userKey(key), data, u.cfg.Secrets.Redis.TTL)
	if err := status.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) ||
			errors.Is(err, context.Canceled) {
			return consts.ErrContext
		}

		return fmt.Errorf("failed to save into cache: %w", err)
	}

	return nil
}

func (u *UserCache) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return u.get(ctx, u.userKey(email))
}

func (u *UserCache) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return u.get(ctx, u.userKey(id))
}

func (u *UserCache) get(ctx context.Context, key string) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result, err := u.client.Get(ctx, u.userKey(key)).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	user, err := u.mapper.ToDomain(result)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) ||
			errors.Is(err, context.Canceled) {
			return nil, consts.ErrContext
		}
		if err == redis.Nil {
			return nil, consts.ErrUserDoesntExist
		}
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return user, nil
}

func (u *UserCache) userKey(key any) string {
	return fmt.Sprintf("user:%s", key)
}

func NewUserCache(cfg *config.Config, client *redis.Client) (*UserCache, error) {
	if client == nil {
		return nil, errors.New("redis client is nil")
	}

	return &UserCache{
		client: client,
		cfg:    cfg,
		mapper: &cache.CacheMapper{},
	}, nil
}
