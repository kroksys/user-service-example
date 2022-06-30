package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	FirstName string
	LastName  string
	Nickname  string
	Password  string
	Email     string `gorm:"unique"`
	Country   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Gorm before create hook is executed before DB.Create()
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
