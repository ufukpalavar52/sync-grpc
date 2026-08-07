package repository_test

import (
	"errors"
	"imapsync-grpc/internal/model"
	"imapsync-grpc/internal/repository"
	"imapsync-grpc/pkg/pb/user"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func strPtr(s string) *string { return &s }

func TestUserRepository_CreateByRequest(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewUserRepository(gdb)

	req := &user.CreateUserRequest{
		Email:        "new@example.com",
		Password:     "hashed-pw",
		Name:         "New User",
		ProfileImage: strPtr("avatar.png"),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	entity, err := repo.CreateByRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity.ID != 1 {
		t.Fatalf("expected ID 1, got %d", entity.ID)
	}
	if entity.Email != "new@example.com" || entity.Name != "New User" {
		t.Fatalf("unexpected entity: %+v", entity)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_CreateByRequest_DBError(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewUserRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(errors.New("duplicate key"))
	mock.ExpectRollback()

	entity, err := repo.CreateByRequest(&user.CreateUserRequest{Email: "dup@example.com"})
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if entity != nil {
		t.Fatalf("expected nil entity on error, got %+v", entity)
	}
}

func TestUserRepository_UpdateByRequest(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewUserRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name"}).AddRow(1, "u@example.com", "Updated Name"))

	entity, err := repo.UpdateByRequest(&user.UpdateUserRequest{Email: "u@example.com", Name: "Updated Name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity.Name != "Updated Name" {
		t.Fatalf("expected updated name, got %+v", entity)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_UpdateByModel(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewUserRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(3, "Renamed"))

	entity, err := repo.UpdateByModel(3, &model.UserEntity{Name: "Renamed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity.Name != "Renamed" {
		t.Fatalf("expected name Renamed, got %+v", entity)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_UpdatePassByRequest(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewUserRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "u@example.com"))

	entity, err := repo.UpdatePassByRequest(&user.UpdateUserPasswordRequest{Email: "u@example.com", NewPassword: "new-hashed-pw"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity.Email != "u@example.com" {
		t.Fatalf("unexpected entity: %+v", entity)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_FindByEmail_Found(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewUserRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "found@example.com"))

	entity, err := repo.FindByEmail("found@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity == nil || entity.Email != "found@example.com" {
		t.Fatalf("unexpected entity: %+v", entity)
	}
}

func TestUserRepository_FindByEmail_NotFoundReturnsNilNil(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewUserRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}))

	entity, err := repo.FindByEmail("missing@example.com")
	if err != nil {
		t.Fatalf("expected nil error for not-found, got %v", err)
	}
	if entity != nil {
		t.Fatalf("expected nil entity for not-found, got %+v", entity)
	}
}

func TestUserRepository_FindById(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewUserRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(9, "nine@example.com"))

	entity, err := repo.FindById(9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity == nil || entity.ID != 9 {
		t.Fatalf("unexpected entity: %+v", entity)
	}
}

func TestUserRepository_FindByFirst_DBError(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewUserRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnError(errors.New("connection refused"))

	entity, err := repo.FindByFirst(model.UserEntity{Email: "x@example.com"})
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if entity != nil {
		t.Fatalf("expected nil entity on error, got %+v", entity)
	}
}
