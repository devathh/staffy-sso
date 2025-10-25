package cache

import "github.com/google/uuid"

type UserModel struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	Password    string
	IsRecruiter bool `json:"is_recruiter"`
}
