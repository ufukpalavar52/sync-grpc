package repository_test

import (
	"errors"
	"imapsync-user/internal/model"
	"imapsync-user/internal/repository"
	"imapsync-user/pkg/pb/sync"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

func TestImapRepository_CreateSync(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewImapRepository(gdb)

	req := &sync.ImapSyncRequest{
		TransactionId:  "tx-1",
		UserId:         5,
		SourceUser:     "src@example.com",
		SourceHost:     "imap.src.example.com",
		SourcePassword: "encrypted-src",
		DestUser:       "dst@example.com",
		DestHost:       "imap.dst.example.com",
		DestPassword:   "encrypted-dst",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "imap_syncs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectCommit()

	entity, err := repo.CreateSync(req, "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity.ID != 10 {
		t.Fatalf("expected ID 10, got %d", entity.ID)
	}
	if entity.TransactionId != "tx-1" || entity.UserID != 5 {
		t.Fatalf("unexpected entity: %+v", entity)
	}
	if entity.Status != "PENDING" {
		t.Fatalf("expected status PENDING, got %s", entity.Status)
	}
	if entity.EncryptionVersion != "v1" {
		t.Fatalf("expected encryption version v1, got %s", entity.EncryptionVersion)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestImapRepository_GetSyncByTransactionId_Found(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewImapRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "imap_syncs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "transaction_id", "status"}).AddRow(1, "tx-1", "PENDING"))

	entity, err := repo.GetSyncByTransactionId("tx-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity == nil || entity.TransactionId != "tx-1" {
		t.Fatalf("unexpected entity: %+v", entity)
	}
}

func TestImapRepository_GetSyncByTransactionId_NotFoundReturnsNilNil(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewImapRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "imap_syncs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "transaction_id"}))

	entity, err := repo.GetSyncByTransactionId("missing-tx")
	if err != nil {
		t.Fatalf("expected nil error for not-found, got %v", err)
	}
	if entity != nil {
		t.Fatalf("expected nil entity for not-found, got %+v", entity)
	}
}

func TestImapRepository_UpdateById_Success(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewImapRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "imap_syncs" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "imap_syncs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(1, "COMPLETED"))

	entity, err := repo.UpdateById(1, &model.ImapSyncEntity{Status: "COMPLETED"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity.Status != "COMPLETED" {
		t.Fatalf("expected status COMPLETED, got %s", entity.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestImapRepository_UpdateById_NoRowsAffectedReturnsNotFound(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewImapRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "imap_syncs" SET`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	entity, err := repo.UpdateById(999, &model.ImapSyncEntity{Status: "COMPLETED"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
	if entity != nil {
		t.Fatalf("expected nil entity, got %+v", entity)
	}
}

func TestImapRepository_FindAllBy_WithPaginationComputesTotal(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewImapRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "imap_syncs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(1, "PENDING"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "imap_syncs"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	result, err := repo.FindAllBy(model.SearchSync{}, model.Pagination{Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GetTotal() != 5 {
		t.Fatalf("expected total 5, got %d", result.GetTotal())
	}
	if len(result.GetData()) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result.GetData()))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestImapRepository_FindAllBy_WithoutPaginationSkipsCount(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewImapRepository(gdb)

	mock.ExpectQuery(`SELECT \* FROM "imap_syncs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(1, "PENDING").AddRow(2, "COMPLETED"))

	result, err := repo.FindAllBy(model.SearchSync{}, model.Pagination{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.GetData()) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(result.GetData()))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (count query should have been skipped): %v", err)
	}
}

func TestImapRepository_GetCountByStatusStats(t *testing.T) {
	gdb, mock := newMockDB(t)
	repo := repository.NewImapRepository(gdb)

	mock.ExpectQuery(`SELECT status, count\(id\) as count, AVG\(finish_time\) as avg_time FROM "imap_syncs" GROUP BY "status"`).
		WillReturnRows(sqlmock.NewRows([]string{"status", "count", "avg_time"}).
			AddRow("PENDING", 3, 12.5).
			AddRow("COMPLETED", 7, 45.2))

	stats, err := repo.GetCountByStatusStats(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stat rows, got %d", len(stats))
	}
	if stats[0].Status != "PENDING" || stats[0].Count != 3 || stats[0].AvgTime != 12.5 {
		t.Fatalf("unexpected first row: %+v", stats[0])
	}
}
