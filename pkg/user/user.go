package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	FirstName string
	LastName  string
	Nickname  string
	Password  string
	Email     string
	Country   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
