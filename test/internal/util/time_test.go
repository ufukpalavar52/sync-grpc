package util_test

import (
	"imapsync-user/internal/util"
	"testing"
	"time"
)

func TestStrToTime_ValidRFC3339(t *testing.T) {
	got, err := util.StrToTime("2026-07-23T10:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestStrToTime_Invalid(t *testing.T) {
	_, err := util.StrToTime("not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date string")
	}
}

func TestStrToTimeWithoutError_Valid(t *testing.T) {
	got := util.StrToTimeWithoutError("2026-07-23T10:00:00Z")
	want := time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestStrToTimeWithoutError_InvalidReturnsZeroTime(t *testing.T) {
	got := util.StrToTimeWithoutError("garbage")
	if !got.IsZero() {
		t.Fatalf("expected zero time for invalid input, got %v", got)
	}
}
