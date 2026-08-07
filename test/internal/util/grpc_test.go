package util_test

import (
	"context"
	"imapsync-user/internal/util"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestGetMetadata_FromIncomingContext(t *testing.T) {
	md := metadata.New(map[string]string{"transaction-id": "abc-123"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	got := util.GetMetadata(ctx)
	if len(got.Get("transaction-id")) == 0 || got.Get("transaction-id")[0] != "abc-123" {
		t.Fatalf("expected transaction-id abc-123, got %v", got.Get("transaction-id"))
	}
}

func TestGetMetadata_NoIncomingMetadata(t *testing.T) {
	got := util.GetMetadata(context.Background())
	if got != nil {
		t.Fatalf("expected nil metadata, got %v", got)
	}
}

func TestGetFirstMetaVal_ReturnsFirstValueAsString(t *testing.T) {
	md := metadata.New(map[string]string{"transaction-id": "abc-123"})

	val := util.GetFirstMetaVal(md, "transaction-id")
	str, ok := val.(string)
	if !ok {
		t.Fatalf("expected a string value, got %T (%v)", val, val)
	}
	if str != "abc-123" {
		t.Fatalf("expected %q, got %q", "abc-123", str)
	}
}

func TestGetFirstMetaVal_MultipleValuesReturnsFirst(t *testing.T) {
	md := metadata.MD{}
	md.Append("transaction-id", "first-val")
	md.Append("transaction-id", "second-val")

	val := util.GetFirstMetaVal(md, "transaction-id")
	if val != "first-val" {
		t.Fatalf("expected %q, got %v", "first-val", val)
	}
}

func TestGetFirstMetaVal_MissingKeyReturnsEmptyString(t *testing.T) {
	md := metadata.New(map[string]string{})

	val := util.GetFirstMetaVal(md, "transaction-id")
	if val != "" {
		t.Fatalf("expected empty string, got %v", val)
	}
}

func TestGetFirstMetaVal_NilMetadataReturnsEmptyString(t *testing.T) {
	val := util.GetFirstMetaVal(nil, "transaction-id")
	if val != "" {
		t.Fatalf("expected empty string, got %v", val)
	}
}

func TestGetMetaTransactionId(t *testing.T) {
	md := metadata.New(map[string]string{util.TransactionIDMetaKey: "tx-42"})

	val := util.GetMetaTransactionId(md)
	if val != "tx-42" {
		t.Fatalf("expected %q, got %v", "tx-42", val)
	}
}
