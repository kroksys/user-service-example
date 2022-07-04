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

func UpdateUserByMap(u *models.User, m map[string]interface{}) error {
	return DB.Model(u).Updates(m).Error
}

func DeleteUser(userId uuid.UUID) error {
	return DB.Delete(&models.User{}, userId).Error
}

func ListUsers(limit, offset int, country string) ([]models.User, error) {
	users := []models.User{}
	if offset > 0 && limit == 0 {
		limit = 1
	}
	tx := DB.Limit(limit).Offset(offset)
	if country != "" {
		tx.Where("country = ?", country)
	}
	err := tx.Find(&users).Error
	return users, err
}
