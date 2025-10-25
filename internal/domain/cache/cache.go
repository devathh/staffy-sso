// Package domain implements cache's domain interface
package domain

import (
	"context"

	domain "github.com/devathh/staffy-sso/internal/domain/user"
	"github.com/google/uuid"
)

type UserCache interface {
	SetByEmail(ctx context.Context, user *domain.User) error
	SetByID(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}
