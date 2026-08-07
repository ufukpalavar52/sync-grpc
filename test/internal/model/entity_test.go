package model_test

import (
	"imapsync-grpc/internal/model"
	"testing"
)

func TestTableNames(t *testing.T) {
	cases := []struct {
		name  string
		table interface{ TableName() string }
		want  string
	}{
		{"UserEntity", &model.UserEntity{}, "users"},
		{"ImapSyncEntity", &model.ImapSyncEntity{}, "imap_syncs"},
		{"EncryptionKeyEntity", &model.EncryptionKeyEntity{}, "encryption_keys"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.table.TableName(); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestEncryptionKeyEntity_GetByteKey(t *testing.T) {
	entity := &model.EncryptionKeyEntity{Key: "aGVsbG8="}

	got := entity.GetByteKey()
	if string(got) != "hello" {
		t.Fatalf("expected %q, got %q", "hello", string(got))
	}
}

func TestEncryptionKeyEntity_GetByteKey_InvalidBase64ReturnsNoBytes(t *testing.T) {
	entity := &model.EncryptionKeyEntity{Key: "not base64!!"}

	got := entity.GetByteKey()
	if len(got) != 0 {
		t.Fatalf("expected no decoded bytes for invalid base64 key, got %v", got)
	}
}
