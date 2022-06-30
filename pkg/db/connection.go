package db

import (
	"github.com/kroksys/user-service-example/pkg/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	// shared database connection
	DB *gorm.DB
)

// Opens the database connection using ConnectionString. Database connection
// can be accessed using globally accessible DB variable.
// It is possible to pass optional Dialecot that's used for testing purposes.
func Connect(connectionString string, d ...gorm.Dialector) error {
	var err error
	dl := mysql.Open(connectionString)
	if len(d) != 0 {
		dl = d[0]
	}
	DB, err = gorm.Open(dl)
	return err
}

// Auto migrates models for database
func Migrate() error {
	return DB.AutoMigrate(&models.User{})
}
