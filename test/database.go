package test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Mysql dialector to connect to gorm mysql mock server.
func MysqlDialector(sqlDB *sql.DB) gorm.Dialector {
	return mysql.New(mysql.Config{
		Conn:       sqlDB,
		DriverName: "mysql",
	})
}

// Get new mock database and opens gorm connection.
func NewMockDatabase(t *testing.T) (*sql.DB, *gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
	)
	require.NoError(t, err, "TestConnect")

	columns := []string{"version"}
	mock.ExpectQuery("SELECT VERSION()").WithArgs().WillReturnRows(
		mock.NewRows(columns).FromCSVString("1"),
	)

	db, err := gorm.Open(MysqlDialector(sqlDB))
	require.NoError(t, err, "NewDatabase")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	return sqlDB, db, mock
}
