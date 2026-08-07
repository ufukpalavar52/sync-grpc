package repository_test

import (
	"imapsync-grpc/internal/model"
	"imapsync-grpc/internal/repository"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestApplySmartFilters_EqualityAndLike(t *testing.T) {
	gdb, _ := newMockDB(t)
	base := repository.NewBaseRepository(gdb)

	search := model.SearchSync{Status: "PENDING", SourceHost: "example"}

	dryRun := base.ApplySmartFilters(&search, true).Session(&gorm.Session{DryRun: true})
	stmt := dryRun.Find(&[]model.ImapSyncEntity{}).Statement

	sql := stmt.SQL.String()
	if !strings.Contains(sql, `status = `) {
		t.Fatalf("expected SQL to filter equality on status, got: %s", sql)
	}
	if !strings.Contains(sql, `source_host LIKE `) {
		t.Fatalf("expected SQL to LIKE-filter on source_host, got: %s", sql)
	}

	foundLikeVal := false
	for _, v := range stmt.Vars {
		if s, ok := v.(string); ok && s == "%example%" {
			foundLikeVal = true
		}
	}
	if !foundLikeVal {
		t.Fatalf("expected LIKE value %%example%%, got vars: %v", stmt.Vars)
	}
}

func TestApplySmartFilters_SkipsZeroValueFields(t *testing.T) {
	gdb, _ := newMockDB(t)
	base := repository.NewBaseRepository(gdb)

	search := model.SearchSync{Status: "PENDING"}

	dryRun := base.ApplySmartFilters(&search, true).Session(&gorm.Session{DryRun: true})
	stmt := dryRun.Find(&[]model.ImapSyncEntity{}).Statement

	sql := stmt.SQL.String()
	if strings.Contains(sql, "transaction_id") {
		t.Fatalf("expected zero-value TransactionId field to be skipped, got SQL: %s", sql)
	}
	if strings.Contains(sql, "source_host") {
		t.Fatalf("expected zero-value SourceHost field to be skipped, got SQL: %s", sql)
	}
}

func TestApplySmartFilters_NoFieldsSetProducesNoWhereClause(t *testing.T) {
	gdb, _ := newMockDB(t)
	base := repository.NewBaseRepository(gdb)

	search := model.SearchSync{}

	dryRun := base.ApplySmartFilters(&search, true).Session(&gorm.Session{DryRun: true})
	stmt := dryRun.Find(&[]model.ImapSyncEntity{}).Statement

	if strings.Contains(stmt.SQL.String(), "WHERE") {
		t.Fatalf("expected no WHERE clause when all fields are zero, got SQL: %s", stmt.SQL.String())
	}
}
