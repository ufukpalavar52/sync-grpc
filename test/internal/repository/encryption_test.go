package repository_test

import (
	"errors"
	"imapsync-user/internal/repository"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

func TestEncryptionRepository_Create(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewEncryptionRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "encryption_keys"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	key, err := repo.Create("plaintext-key", "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.ID != 1 {
		t.Fatalf("expected ID 1, got %d", key.ID)
	}
	if key.Key != "plaintext-key" || key.Version != "v1" {
		t.Fatalf("unexpected key: %+v", key)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestEncryptionRepository_Create_DBError(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewEncryptionRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "encryption_keys"`).
		WillReturnError(errors.New("connection lost"))
	mock.ExpectRollback()

	key, err := repo.Create("plaintext-key", "v1")
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if key != nil {
		t.Fatalf("expected nil key on error, got %+v", key)
	}
}

func TestEncryptionRepository_FindByVersion_Found(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewEncryptionRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "encryption_keys"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "version"}).AddRow(1, "encoded-key", "v1"))

	key, err := repo.FindByVersion("v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected a non-nil key")
	}
	if key.Version != "v1" || key.Key != "encoded-key" {
		t.Fatalf("unexpected key: %+v", key)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestEncryptionRepository_FindByVersion_NotFoundReturnsNilNil(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewEncryptionRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "encryption_keys"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "version"}))

	key, err := repo.FindByVersion("missing-version")
	if err != nil {
		t.Fatalf("expected nil error for not-found, got %v", err)
	}
	if key != nil {
		t.Fatalf("expected nil key for not-found, got %+v", key)
	}
}

func TestEncryptionRepository_FindByVersion_DBError(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewEncryptionRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "encryption_keys"`).
		WillReturnError(gorm.ErrInvalidDB)

	key, err := repo.FindByVersion("v1")
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if key != nil {
		t.Fatalf("expected nil key on error, got %+v", key)
	}
}
