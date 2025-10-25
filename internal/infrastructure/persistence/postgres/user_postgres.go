package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	domain "github.com/devathh/staffy-sso/internal/domain/user"
	"github.com/devathh/staffy-sso/internal/infrastructure/persistence"
	"github.com/devathh/staffy-sso/pkg/consts"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type userRepository struct {
	db          *gorm.DB
	pgDialector postgres.Dialector
	mapper      persistence.UserMapper
}

func (ur *userRepository) Save(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	if err := ctx.Err(); err != nil {
		return uuid.Nil, err
	}

	if user == nil {
		return uuid.Nil, consts.ErrEmptyUser
	}

	userModel, err := ur.mapper.ToModel(user)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to convert from domain to model: %w", err)
	}

	if err := ur.db.WithContext(ctx).Create(&userModel).Error; err != nil {
		if errors.Is(ur.pgDialector.Translate(err), gorm.ErrDuplicatedKey) {
			return uuid.Nil, consts.ErrUserAlreadyExists
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return uuid.Nil, consts.ErrContext
		}

		return uuid.Nil, fmt.Errorf("failed to save user: %w", err)
	}

	return userModel.ID, nil
}

func (ur *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	result := ur.db.WithContext(ctx).Delete(&persistence.UserModel{}, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, context.DeadlineExceeded) ||
			errors.Is(result.Error, context.Canceled) {
			return consts.ErrContext
		}

		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return consts.ErrUserDoesntExist
	}

	return nil
}

func (ur *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var userModel persistence.UserModel
	if err := ur.db.WithContext(ctx).First(&userModel, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	user, err := ur.mapper.ToDomain(&userModel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert from model to domain: %w", err)
	}

	return user, nil
}

func (ur *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	var userModel persistence.UserModel
	if err := ur.db.WithContext(ctx).First(&userModel, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user, err := ur.mapper.ToDomain(&userModel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert from model to domain: %w", err)
	}

	return user, nil
}

func NewUserRepository(db *gorm.DB) (domain.UserRepository, error) {
	if db == nil {
		return nil, errors.New("db cannot be empty")
	}

	return &userRepository{
		pgDialector: postgres.Dialector{},
		db:          db,
		mapper:      persistence.UserMapper{},
	}, nil
}
