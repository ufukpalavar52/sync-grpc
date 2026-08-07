package util

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var UserNotFound = status.Error(codes.NotFound, "user not found")
var InvalidEmailOrPassword = status.Error(codes.NotFound, "invalid email or password")
var UserAlreadyExists = status.Error(codes.AlreadyExists, "user already exists")
var SyncAlreadyExists = status.Error(codes.AlreadyExists, "sync with the same transaction id already exists")
var SyncNotFound = status.Error(codes.NotFound, "sync not found")
var KeyVersionNotFound = status.Error(codes.NotFound, "key version not found")

func UnknownError(err error) error {
	return status.Error(codes.Unknown, err.Error())
}

func InvalidArgument(err error) error {
	return status.Error(codes.InvalidArgument, err.Error())
}
