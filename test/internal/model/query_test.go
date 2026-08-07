package model_test

import (
	"imapsync-grpc/internal/model"
	"imapsync-grpc/pkg/pb/sync"
	"testing"
)

func TestNewSearchSyncFromReq(t *testing.T) {
	req := &sync.ListSyncRequest{
		UserId:        7,
		TransactionId: "tx-1",
		SourceHost:    "src.example.com",
		SourceUser:    "srcUser",
		DestHost:      "dst.example.com",
		DestUser:      "dstUser",
		Status:        "PENDING",
	}

	search := model.NewSearchSyncFromReq(req)

	if search.UserId != 7 {
		t.Fatalf("expected UserId 7, got %d", search.UserId)
	}
	if search.TransactionId != "tx-1" {
		t.Fatalf("expected TransactionId tx-1, got %s", search.TransactionId)
	}
	if search.SourceHost != "src.example.com" {
		t.Fatalf("expected SourceHost src.example.com, got %s", search.SourceHost)
	}
	if search.SourceUser != "srcUser" {
		t.Fatalf("expected SourceUser srcUser, got %s", search.SourceUser)
	}
	if search.DestHost != "dst.example.com" {
		t.Fatalf("expected DestHost dst.example.com, got %s", search.DestHost)
	}
	if search.DestUser != "dstUser" {
		t.Fatalf("expected DestUser dstUser, got %s", search.DestUser)
	}
	if search.Status != "PENDING" {
		t.Fatalf("expected Status PENDING, got %s", search.Status)
	}
}

func TestPaginationResult(t *testing.T) {
	result := model.NewPaginationResult[[]string](42, []string{"a", "b"})

	if result.GetTotal() != 42 {
		t.Fatalf("expected total 42, got %d", result.GetTotal())
	}
	if len(result.GetData()) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.GetData()))
	}
}
