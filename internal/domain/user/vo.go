package domain

import (
	"fmt"
	"net/mail"
	"strings"
)

type Email struct {
	email string
}

func NewEmail(email string) (Email, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return Email{}, fmt.Errorf("invalid email: %w", err)
	}

	return Email{
		email: strings.ToLower(email),
	}, nil
}
