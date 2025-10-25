// Package consts is just a set of constrains and errors
package consts

import "errors"

var (
	ErrEmptyUser         = errors.New("user cannot be empty")
	ErrUserDoesntExist   = errors.New("user doesn't exist")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidEmail      = errors.New("email is invalid")
	ErrCreateUser        = errors.New("failed to create user")

	ErrContext  = errors.New("context was canceled or is timeout")
	ErrDatabase = errors.New("error with database")

	ErrNilToken           = errors.New("token cannot be nil")
	ErrGenerateToken      = errors.New("failed to generate new token")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrNilRequest  = errors.New("request cannot be nil")
	ErrInvalidArgs = errors.New("some args is invalid")

	ErrNilCfg = errors.New("cfg cannot be nil")
)
