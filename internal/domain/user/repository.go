package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Save(context.Context, *User) (uuid.UUID, error)
	Delete(context.Context, uuid.UUID) error
	GetByID(context.Context, uuid.UUID) (*User, error)
	GetByEmail(context.Context, string) (*User, error)
}
