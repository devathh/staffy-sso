// Package services implements a service layer of application
package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	staffy "github.com/devathh/staffy-proto/gen/go"
	domainCache "github.com/devathh/staffy-sso/internal/domain/cache"
	"github.com/devathh/staffy-sso/internal/domain/observability"
	domain "github.com/devathh/staffy-sso/internal/domain/user"
	"github.com/devathh/staffy-sso/internal/infrastructure/config"
	"github.com/devathh/staffy-sso/internal/lib/jwt"
	"github.com/devathh/staffy-sso/pkg/consts"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

type ssoService struct {
	log         *slog.Logger
	persistence domain.UserRepository
	cache       domainCache.UserCache
	jwt         *jwt.JWT
	cfg         *config.Config
	ch          observability.UserCH
}

type SSOService interface {
	GetUserByToken(ctx context.Context, token *staffy.Token) (*staffy.User, error)
	Login(ctx context.Context, req *staffy.LoginRequest) (*staffy.AuthResponse, error)
	Register(ctx context.Context, req *staffy.RegisterRequest) (*staffy.AuthResponse, error)
	Delete(ctx context.Context, token *staffy.Token) (*staffy.StatusResponse, error)
	Refresh(ctx context.Context, token *staffy.Token) (*staffy.Token, error)
}

func (s *ssoService) GetUserByToken(ctx context.Context, token *staffy.Token) (*staffy.User, error) {
	if token == nil {
		return nil, consts.ErrNilToken
	}

	start := time.Now().UTC()

	claims, err := s.getClaimsFromToken(token)
	if err != nil {
		return nil, err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.RWTimeout)
	defer cancel()

	// The first step is trying to get the user from the cache
	user, err := s.getUserFromCacheByID(ctxTimeout, claims.ID)
	if err == nil {
		// Validate given user n' token user
		if user.Email() != claims.Email {
			return nil, consts.ErrInvalidToken
		}

		go s.saveLog(context.TODO(), staffy.SSO_GetUserByToken_FullMethodName, time.Since(start), int(codes.OK), true)
		return s.toStaffyUser(user), nil
	}

	// If it didn't work out, we try to transfer to the database
	user, err = s.persistence.GetByID(ctxTimeout, claims.ID)
	if err != nil {
		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, consts.ErrUserDoesntExist
		}

		s.log.Error("failed to get user by id", slog.String("error", err.Error()),
			slog.String("user_id", claims.ID.String()),
			slog.String("user_email", claims.Email))
		return nil, consts.ErrDatabase
	}

	// Validate given user n' token user
	if user.Email() != claims.Email {
		return nil, consts.ErrInvalidToken
	}

	// Try to save user to cache
	go func() {
		if err := s.saveUserToCacheByID(user); err != nil {
			s.log.Error("failed to save user into cache", slog.String("error", err.Error()))
		}
	}()

	go s.saveLog(context.TODO(), staffy.SSO_GetUserByToken_FullMethodName, time.Since(start), int(codes.OK), false)
	return s.toStaffyUser(user), nil
}

func (s *ssoService) Login(ctx context.Context, req *staffy.LoginRequest) (*staffy.AuthResponse, error) {
	if req == nil {
		return nil, consts.ErrNilRequest
	}

	start := time.Now().UTC()

	// Converting arguments to normal form
	email, password := strings.TrimSpace(req.GetEmail()), strings.TrimSpace(req.GetPassword())
	if email == "" ||
		password == "" {
		return nil, consts.ErrInvalidArgs
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.RWTimeout)
	defer cancel()

	// At first, try to get user from cache by email
	user, err := s.getUserFromCacheByEmail(ctxTimeout, email)
	if err == nil {
		if !user.CheckThePassword(password) {
			return nil, consts.ErrInvalidCredentials
		}

		go s.saveLog(context.TODO(), staffy.SSO_Login_FullMethodName, time.Since(start), int(codes.OK), true)
		return s.toAuthResponse(user)
	}

	// If didn't work out - try to get from db
	user, err = s.persistence.GetByEmail(ctxTimeout, email)
	if err != nil {
		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, consts.ErrInvalidCredentials
		}

		s.log.Error("failed to get user by email", slog.String("error", err.Error()))
		return nil, consts.ErrDatabase
	}

	if !user.CheckThePassword(password) {
		return nil, consts.ErrInvalidCredentials
	}

	// Save this user to cache
	go func() {
		if err := s.saveUserToCacheByEmail(user); err != nil {
			s.log.Error("failed to save user into cache", slog.String("error", err.Error()))
		}
	}()

	go s.saveLog(context.TODO(), staffy.SSO_Login_FullMethodName, time.Since(start), int(codes.OK), false)
	return s.toAuthResponse(user)
}

func (s *ssoService) Register(ctx context.Context, req *staffy.RegisterRequest) (*staffy.AuthResponse, error) {
	if req == nil {
		return nil, consts.ErrNilRequest
	}

	start := time.Now().UTC()

	email, err := domain.NewEmail(strings.TrimSpace(req.GetEmail()))
	if err != nil {
		return nil, consts.ErrInvalidEmail
	}

	user, err := domain.NewUser(email, req.GetName(), req.GetSurname(), req.GetPassword(), req.GetIsRecruiter())
	if err != nil {
		s.log.Error("failed to create new user", slog.String("error", err.Error()))
		return nil, consts.ErrCreateUser
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.RWTimeout)
	defer cancel()

	// Save user to db
	id, err := s.persistence.Save(ctxTimeout, user)
	if err != nil {
		if errors.Is(err, consts.ErrUserAlreadyExists) {
			return nil, consts.ErrUserAlreadyExists
		}

		s.log.Error("failed to save user", slog.String("error", err.Error()))
		return nil, consts.ErrDatabase
	}

	go s.saveLog(context.TODO(), staffy.SSO_Register_FullMethodName, time.Since(start), int(codes.OK), false)
	return s.toAuthResponse(
		domain.FromPersistence(id,
			email,
			user.Name(),
			user.Surname(),
			user.Password(),
			user.IsRecruiter(),
		),
	)
}

func (s *ssoService) Delete(ctx context.Context, token *staffy.Token) (*staffy.StatusResponse, error) {
	if token == nil {
		return nil, consts.ErrNilToken
	}

	start := time.Now().UTC()

	tokenString := strings.TrimSpace(token.GetToken())
	if tokenString == "" {
		return nil, consts.ErrNilToken
	}

	claims, err := s.jwt.ValidateToken(tokenString)
	if err != nil {
		s.log.Warn("invalid token detected", slog.String("error", err.Error()))
		return nil, consts.ErrInvalidToken
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.RWTimeout)
	defer cancel()

	if err := s.persistence.Delete(ctxTimeout, claims.ID); err != nil {
		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, consts.ErrUserDoesntExist
		}

		s.log.Error("failed to delete user", slog.String("error", err.Error()))
		return nil, consts.ErrDatabase
	}

	go s.saveLog(context.TODO(), staffy.SSO_Delete_FullMethodName, time.Since(start), int(codes.OK), false)
	return &staffy.StatusResponse{
		Timestamp:     time.Now().UTC().Unix(),
		StatusCode:    http.StatusOK,
		StatusMessage: "user has been deleted",
	}, nil
}

func (s *ssoService) Refresh(ctx context.Context, token *staffy.Token) (*staffy.Token, error) {
	if token == nil {
		return nil, consts.ErrNilToken
	}

	start := time.Now().UTC()

	tokenString := strings.TrimSpace(token.GetToken())
	if tokenString == "" {
		return nil, consts.ErrNilToken
	}

	claims, err := s.jwt.ValidateToken(tokenString)
	if err != nil {
		s.log.Warn("invalid token detected", slog.String("error", err.Error()))
		return nil, consts.ErrInvalidToken
	}

	newToken, err := s.jwt.GenerateToken(claims.Email, claims.ID)
	if err != nil {
		s.log.Error("failed to generate new token", slog.String("error", err.Error()))
		return nil, consts.ErrGenerateToken
	}

	go s.saveLog(context.TODO(), staffy.SSO_Refresh_FullMethodName, time.Since(start), int(codes.OK), false)
	return &staffy.Token{
		Token: newToken,
	}, nil
}

func (s *ssoService) toAuthResponse(user *domain.User) (*staffy.AuthResponse, error) {
	tokenString, err := s.jwt.GenerateToken(user.Email(), user.ID())
	if err != nil {
		s.log.Error("failed to generate new token", slog.String("error", err.Error()))
		return nil, consts.ErrGenerateToken
	}

	return &staffy.AuthResponse{
		Token: tokenString,
		User:  s.toStaffyUser(user),
	}, nil
}

func (s *ssoService) toStaffyUser(user *domain.User) *staffy.User {
	return &staffy.User{
		UserId:      user.ID().String(),
		Email:       user.Email(),
		Name:        user.Name(),
		Surname:     user.Surname(),
		IsRecruiter: user.IsRecruiter(),
	}
}

func (s *ssoService) getUserFromCacheByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.cache.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	return user, nil
}

func (s *ssoService) getUserFromCacheByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.cache.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	return user, nil
}

func (s *ssoService) saveUserToCacheByID(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.RWTimeout)
	defer cancel()

	if err := s.cache.SetByID(ctx, user); err != nil {
		return fmt.Errorf("failed to save user to cache: %w", err)
	}

	return nil
}

func (s *ssoService) saveUserToCacheByEmail(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.RWTimeout)
	defer cancel()

	if err := s.cache.SetByEmail(ctx, user); err != nil {
		return fmt.Errorf("failed to save user to cache: %w", err)
	}

	return nil
}

func (s *ssoService) getClaimsFromToken(token *staffy.Token) (*jwt.CustomClaims, error) {
	tokenString := strings.TrimSpace(token.GetToken())
	if tokenString == "" {
		return nil, consts.ErrNilToken
	}

	claims, err := s.jwt.ValidateToken(tokenString)
	if err != nil {
		s.log.Warn("invalid token detected", slog.String("error", err.Error()))
		return nil, consts.ErrInvalidToken
	}

	return claims, nil
}

func (s *ssoService) saveLog(ctx context.Context, endpoint string, duration time.Duration, code int, cacheHit bool) {
	performanceLog := observability.PerformanceLog{
		Endpoint:   endpoint,
		Duration:   duration,
		StatusCode: code,
		CacheHit:   cacheHit,
	}
	s.ch.SavePerformanceLog(ctx, &performanceLog)
}

func NewSSOService(cfg *config.Config, log *slog.Logger, persistence domain.UserRepository, cache domainCache.UserCache, ch observability.UserCH, jwt *jwt.JWT) SSOService {
	return &ssoService{
		log:         log,
		persistence: persistence,
		cache:       cache,
		jwt:         jwt,
		cfg:         cfg,
		ch:          ch,
	}
}
