// Package domain implements user's domain structure
package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	id          uuid.UUID
	email       Email
	password    string
	name        string
	surname     string
	isRecruiter bool
}

func (u *User) CheckThePassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(u.password), []byte(password)); err != nil {
		return false
	}

	return true
}

func (u User) Password() string {
	return u.password
}

func (u User) ID() uuid.UUID {
	return u.id
}

func (u User) Email() string {
	return u.email.email
}

func (u User) Name() string {
	return u.name
}

func (u User) Surname() string {
	return u.surname
}

func (u User) IsRecruiter() bool {
	return u.isRecruiter
}

func (u *User) ToRecruiter() error {
	if u.isRecruiter {
		return errors.New("user is already recruiter")
	}
	u.isRecruiter = true

	return nil
}

func (u *User) ToApplicant() error {
	if !u.isRecruiter {
		return errors.New("user is already applicant")
	}
	u.isRecruiter = false

	return nil
}

func NewUser(email Email, name, surname, password string, isRecruiter bool) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash password: %w", err)
	}

	// Surname can be empty
	surname = strings.TrimSpace(surname)
	return &User{
		id:          uuid.New(),
		email:       email,
		name:        name,
		surname:     surname,
		isRecruiter: isRecruiter,
		password:    string(passwordHash),
	}, nil
}

func FromPersistence(id uuid.UUID, email Email, name, surname, password string, isRecruiter bool) *User {
	return &User{
		id:          id,
		email:       email,
		name:        name,
		surname:     surname,
		isRecruiter: isRecruiter,
		password:    password,
	}
}
