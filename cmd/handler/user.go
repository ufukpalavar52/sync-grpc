package handler

import (
	"context"
	"imapsync-grpc/internal/model"
	"imapsync-grpc/internal/service"
	"imapsync-grpc/internal/util"
	"imapsync-grpc/logger"
	"imapsync-grpc/pkg/pb/user"
	"log"
	"time"

	"buf.build/go/protovalidate"
	"gorm.io/gorm"
)

type UserServer struct {
	user.UnimplementedUserServiceServer
	validator   protovalidate.Validator
	userService *service.UserService
}

func NewUserServer(db *gorm.DB) *UserServer {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}
	return &UserServer{validator: v, userService: service.NewUserService(db)}
}

func (s *UserServer) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.UserResponse, error) {
	defer logger.InfoL("CreateUser request was finished", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}
	create, err := s.userService.CreateUser(ctx, req)

	return s.transformUserResponse(create, err)
}

func (s *UserServer) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UserResponse, error) {
	defer logger.InfoL("UpdateUser request was finished", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}
	create, err := s.userService.UpdateUser(ctx, req)
	return s.transformUserResponse(create, err)
}

func (s *UserServer) UpdateUserPassword(ctx context.Context, req *user.UpdateUserPasswordRequest) (*user.UserResponse, error) {
	defer logger.InfoL("UpdateUserPassword request was finished", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}
	create, err := s.userService.UpdateUserPassword(ctx, req)
	return s.transformUserResponse(create, err)
}

func (s *UserServer) ResetUserPassword(ctx context.Context, req *user.ResetUserPasswordRequest) (*user.UserResponse, error) {
	defer logger.InfoL("UpdateUserPassword request was finished", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}
	create, err := s.userService.ResetUserPassword(ctx, req)
	return s.transformUserResponse(create, err)
}

func (s *UserServer) AuthUser(ctx context.Context, req *user.AuthUserRequest) (*user.UserResponse, error) {
	defer logger.InfoL("AuthUser request was finished.", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}

	userEntity, err := s.userService.AuthUser(ctx, req)
	return s.transformUserResponse(userEntity, err)
}

func (s *UserServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.UserResponse, error) {
	defer logger.InfoL("AuthUser request was finished.", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}

	userEntity, err := s.userService.GetUser(ctx, req)
	return s.transformUserResponse(userEntity, err)
}

func (s *UserServer) transformUserResponse(entity *model.UserEntity, err error) (*user.UserResponse, error) {
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, util.SyncNotFound
	}

	return &user.UserResponse{
		Id:                entity.ID,
		Email:             entity.Email,
		Name:              entity.Name,
		ProfileImage:      entity.ProfileImage,
		CreatedAt:         entity.CreatedAt.Format(time.DateTime),
		UpdatedAt:         entity.UpdatedAt.Format(time.DateTime),
		EmailVerifiedTime: entity.EmailVerifiedTime,
	}, nil
}
