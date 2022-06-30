package db

import (
	"github.com/google/uuid"
	"github.com/kroksys/user-service-example/pkg/models"
)

func CreateUser(u *models.User) error {
	return DB.Create(u).Error
}

func UpdateUser(u *models.User) error {
	return DB.Save(u).Error
}

func DeleteUser(userId uuid.UUID) error {
	return DB.Delete(&models.User{}, userId).Error
}

func ListUsers(limit, offset int, country string) ([]models.User, error) {
	users := []models.User{}
	tx := DB.Limit(limit).Offset(offset)
	if country != "" {
		tx.Where("country = ?", country)
	}
	return users, tx.Find(&users).Error
}
