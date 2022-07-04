package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/kroksys/user-service-example/pkg/pb/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (u *User) ToUserResponse() *pb.UserResponse {
	return &pb.UserResponse{
		Id:        u.ID.String(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		Password:  u.Password,
		Email:     u.Email,
		Country:   u.Country,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

// Gorm before create hook is executed before DB.Create()
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
