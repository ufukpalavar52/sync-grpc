package util_test

import (
	"errors"
	"imapsync-grpc/internal/util"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPredefinedErrors_Codes(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code codes.Code
	}{
		{"UserNotFound", util.UserNotFound, codes.NotFound},
		{"InvalidEmailOrPassword", util.InvalidEmailOrPassword, codes.NotFound},
		{"UserAlreadyExists", util.UserAlreadyExists, codes.AlreadyExists},
		{"SyncAlreadyExists", util.SyncAlreadyExists, codes.AlreadyExists},
		{"SyncNotFound", util.SyncNotFound, codes.NotFound},
		{"KeyVersionNotFound", util.KeyVersionNotFound, codes.NotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := status.Code(tc.err); got != tc.code {
				t.Fatalf("expected code %v, got %v", tc.code, got)
			}
		})
	}
}

func TestUnknownError(t *testing.T) {
	inner := errors.New("boom")
	err := util.UnknownError(inner)

	if status.Code(err) != codes.Unknown {
		t.Fatalf("expected codes.Unknown, got %v", status.Code(err))
	}
	if status.Convert(err).Message() != "boom" {
		t.Fatalf("expected message %q, got %q", "boom", status.Convert(err).Message())
	}
}

func TestInvalidArgument(t *testing.T) {
	inner := errors.New("bad input")
	err := util.InvalidArgument(inner)

	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected codes.InvalidArgument, got %v", status.Code(err))
	}
	if status.Convert(err).Message() != "bad input" {
		t.Fatalf("expected message %q, got %q", "bad input", status.Convert(err).Message())
	}
}
