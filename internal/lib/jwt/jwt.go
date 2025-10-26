// Package jwt implements structure, that generates n' validates jwt-token
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/devathh/staffy-sso/internal/infrastructure/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CustomClaims struct {
	Email string
	ID    uuid.UUID
	jwt.RegisteredClaims
}

type JWT struct {
	cfg *config.Config
}

func (j *JWT) GenerateToken(email string, id uuid.UUID) (string, error) {
	secretKey := []byte(j.cfg.Secrets.JWT.SecretKey)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &CustomClaims{
		Email: email,
		ID:    id,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "staffy",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.cfg.Secrets.JWT.TTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   id.String(),
		},
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to get string of token: %w", err)
	}

	return tokenString, nil
}

func (j *JWT) ValidateToken(tokenString string) (*CustomClaims, error) {
	secretKey := []byte(j.cfg.Secrets.JWT.SecretKey)

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token")
		}

		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func NewJWT(cfg *config.Config) *JWT {
	return &JWT{
		cfg: cfg,
	}
}
