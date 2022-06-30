package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kroksys/user-service-example/test"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
	)
	require.NoError(t, err, "TestConnect")

	columns := []string{"version"}
	mock.ExpectQuery("SELECT VERSION()").WithArgs().WillReturnRows(
		mock.NewRows(columns).FromCSVString("1"),
	)

	err = Connect("", test.MysqlDialector(sqlDB))
	require.NoError(t, err, "NewDatabase")

	require.NotNil(t, DB)

	mock.ExpectClose()
	sqlDB.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestConnect: %s", err)
	}
}

func TestMigrate(t *testing.T) {
	sqlDB, gdb, mock := test.NewMockDatabase(t)
	defer sqlDB.Close()

	DB = gdb
	require.NotNil(t, DB)

	mock.ExpectQuery("SELECT DATABASE()")
	mock.ExpectQuery("SELECT SCHEMA_NAME from Information_schema.SCHEMATA")
	Migrate()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestMigrate: %s", err)
	}
}
