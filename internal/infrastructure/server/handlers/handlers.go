package handlers

import (
	"context"
	"errors"

	staffy "github.com/devathh/staffy-proto/gen/go"
	"github.com/devathh/staffy-sso/internal/application/services"
	"github.com/devathh/staffy-sso/pkg/consts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SSOHandlers struct {
	service services.SSOService

	staffy.UnimplementedSSOServer
}

func (h *SSOHandlers) GetUserByToken(ctx context.Context, token *staffy.Token) (*staffy.User, error) {
	if token == nil {
		return nil, status.Error(codes.InvalidArgument, "token cannot be empty")
	}

	user, err := h.service.GetUserByToken(ctx, token)
	if err != nil {
		if errors.Is(err, consts.ErrInvalidToken) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return user, nil
}

func (h *SSOHandlers) Login(ctx context.Context, req *staffy.LoginRequest) (*staffy.AuthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := h.service.Login(ctx, req)
	if err != nil {
		if errors.Is(err, consts.ErrInvalidArgs) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, consts.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (h *SSOHandlers) Register(ctx context.Context, req *staffy.RegisterRequest) (*staffy.AuthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := h.service.Register(ctx, req)
	if err != nil {
		if errors.Is(err, consts.ErrCreateUser) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, consts.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (h *SSOHandlers) Delete(ctx context.Context, token *staffy.Token) (*staffy.StatusResponse, error) {
	if token == nil {
		return nil, status.Error(codes.InvalidArgument, "token cannot be empty")
	}

	resp, err := h.service.Delete(ctx, token)
	if err != nil {
		if errors.Is(err, consts.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, status.Error(codes.Unauthenticated, "token is invalid")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (h *SSOHandlers) Refresh(ctx context.Context, token *staffy.Token) (*staffy.Token, error) {
	if token == nil {
		return nil, status.Error(codes.InvalidArgument, "token cannot be empty")
	}

	resp, err := h.service.Refresh(ctx, token)
	if err != nil {
		if errors.Is(err, consts.ErrInvalidToken) {
			return nil, status.Error(codes.InvalidArgument, "token is invalid")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, err
}

func NewHandler(service services.SSOService) *SSOHandlers {
	return &SSOHandlers{
		service: service,
	}
}
