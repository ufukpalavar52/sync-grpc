package util_test

import (
	"imapsync-grpc/internal/util"
	"testing"
)

func TestHashPassword_ProducesVerifiableHash(t *testing.T) {
	hash, err := util.HashPassword("s3cret!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" || hash == "s3cret!" {
		t.Fatalf("expected a bcrypt hash distinct from the plaintext, got %q", hash)
	}
	if !util.CheckPasswordHash("s3cret!", hash) {
		t.Fatal("expected password to match its own hash")
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	hash, err := util.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if util.CheckPasswordHash("wrong-password", hash) {
		t.Fatal("expected wrong password to not match")
	}
}

func TestHashPassword_DifferentSaltEachTime(t *testing.T) {
	h1, err := util.HashPassword("same-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h2, err := util.HashPassword("same-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h1 == h2 {
		t.Fatal("expected different hashes for the same password due to random salt")
	}
}
