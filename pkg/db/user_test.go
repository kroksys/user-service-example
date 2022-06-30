package db

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/kroksys/user-service-example/pkg/models"
	"github.com/kroksys/user-service-example/test"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	sqlDB, gdb, mock := test.NewMockDatabase(t)
	DB = gdb
	defer sqlDB.Close()

	rec := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Nickname:  "Johny",
		Password:  "secret",
		Email:     "johndoe@email.com",
		Country:   "UK",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT`)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := CreateUser(&rec)
	require.NoError(t, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreateUser: %s", err)
	}

	if rec.ID == uuid.Nil {
		t.Errorf("TestCreateUser: user.ID should not be empty after creating user")
	}

	if rec.UpdatedAt.IsZero() {
		t.Errorf("TestCreateUser: user.UpdatedAt should not be empty after creating user")
	}

	if rec.CreatedAt.IsZero() {
		t.Errorf("TestCreateUser: user.CreatedAt should not be empty after creating user")
	}
}

func TestUpdateUser(t *testing.T) {
	sqlDB, gdb, mock := test.NewMockDatabase(t)
	DB = gdb
	defer sqlDB.Close()

	id, _ := uuid.Parse("cc9b61e3-0cba-473f-8e95-944661c46051")
	yearBefore := time.Now().AddDate(-1, 0, 0)
	rec := models.User{
		ID:        id,
		FirstName: "John",
		LastName:  "Doe",
		Nickname:  "Johny",
		Password:  "secret",
		Email:     "johndoe@email.com",
		Country:   "UK",
		UpdatedAt: yearBefore,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE`)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := UpdateUser(&rec)
	require.NoError(t, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestUpdateUser: %s", err)
	}

	if rec.UpdatedAt.Equal(yearBefore) {
		t.Errorf("TestUpdateUser: user.UpdatedAt was not set after updating user record")
	}

	// Update user with emtpy ID should return error
	err = UpdateUser(&models.User{})
	require.Error(t, err)
}

func TestDeleteUser(t *testing.T) {
	sqlDB, gdb, mock := test.NewMockDatabase(t)
	DB = gdb
	defer sqlDB.Close()

	id, _ := uuid.Parse("cc9b61e3-0cba-473f-8e95-944661c46051")
	rec := models.User{
		ID: id,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := DeleteUser(rec.ID)
	require.NoError(t, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestDeleteUser: %s", err)
	}
}

func TestListUser(t *testing.T) {
	sqlDB, gdb, mock := test.NewMockDatabase(t)
	DB = gdb
	defer sqlDB.Close()

	rs := mock.NewRows([]string{"id", "first_name", "last_name", "nickname", "password", "email", "country", "updated_at", "created_at"}).
		AddRow(uuid.New(), "John", "Doe", "johny", "secret", "johndoe@email.com", "UK", time.Now(), time.Now()).
		AddRow(uuid.New(), "John1", "Doe", "johny", "secret", "john1doe@email.com", "UK", time.Now(), time.Now()).
		AddRow(uuid.New(), "John2", "Doe", "johny", "secret", "john2doe@email.com", "UK", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(rs)

	users, err := ListUsers(0, 0, "UK")
	require.NoError(t, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestListUser: %s", err)
	}

	if len(users) != 3 {
		t.Error("TestListUser: found less users than provided")
	}
}
