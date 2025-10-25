package persistence

import "github.com/google/uuid"

type UserModel struct {
	ID          uuid.UUID `gorm:"primarykey"`
	Email       string    `gorm:"unique"`
	Name        string    `gorm:"not null"`
	Surname     string
	IsRecruiter bool
	Password    string
}
