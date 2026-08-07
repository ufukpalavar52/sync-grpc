package service

import (
	"context"
	"imapsync-user/internal/model"
	"imapsync-user/internal/repository"
	"imapsync-user/internal/util"
	"imapsync-user/logger"
	"imapsync-user/pkg/pb/user"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{userRepo: repository.NewUserRepository(db)}
}

func (s *UserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*model.UserEntity, error) {
	userEntity, err := s.userRepo.FindByEmail(req.GetEmail())
	if err != nil {
		return nil, util.UnknownError(err)
	}

	if userEntity != nil {
		return nil, util.UserAlreadyExists
	}

	req.Password, err = util.HashPassword(req.Password)
	if err != nil {
		logger.ErrorL("Failed to hash password",
			util.GetMetadata(ctx),
			logger.DT{"email": req.GetEmail(), "err": err.Error()},
		)
		return nil, util.UnknownError(err)
	}

	create, err := s.userRepo.CreateByRequest(req)

	if err != nil {
		logger.ErrorL("Failed to create user",
			util.GetMetadata(ctx),
			logger.DT{"email": req.GetEmail(), "err": err.Error()},
		)
		return nil, util.UnknownError(err)
	}
	logger.InfoL("Created user", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
	return create, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*model.UserEntity, error) {
	userEntity, err := s.userRepo.FindByEmail(req.GetEmail())
	if err != nil {
		logger.ErrorL("User find error", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
		return nil, util.UnknownError(err)
	}

	if userEntity == nil {
		logger.InfoL("User not found by email", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
		return nil, util.UserNotFound
	}

	userModel := s.mapToEntityByUpdReq(req)

	update, err := s.userRepo.UpdateByModel(userEntity.ID, userModel)
	if err != nil {
		logger.ErrorL("Failed to update user", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
		return nil, util.UnknownError(err)
	}

	return update, nil
}

func (s *UserService) UpdateUserPassword(ctx context.Context, req *user.UpdateUserPasswordRequest) (*model.UserEntity, error) {
	authReq := &user.AuthUserRequest{
		Password: req.GetPassword(),
		Email:    req.GetEmail(),
	}

	_, err := s.AuthUser(ctx, authReq)
	if err != nil {
		return nil, err
	}

	req.NewPassword, err = util.HashPassword(req.NewPassword)
	if err != nil {
		return nil, util.UnknownError(err)
	}

	userEntity, err := s.userRepo.UpdatePassByRequest(req)
	if err != nil {
		return nil, util.UnknownError(err)
	}

	return userEntity, nil
}

func (s *UserService) ResetUserPassword(ctx context.Context, req *user.ResetUserPasswordRequest) (*model.UserEntity, error) {
	_, err := s.CheckUser(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}

	req.Password, err = util.HashPassword(req.GetPassword())
	if err != nil {
		logger.ErrorL("Unknown error", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
		return nil, util.UnknownError(err)
	}

	newReq := &user.UpdateUserPasswordRequest{Email: req.GetEmail(), NewPassword: req.GetPassword()}
	userEntity, err := s.userRepo.UpdatePassByRequest(newReq)
	if err != nil {
		return nil, util.UnknownError(err)
	}

	return userEntity, nil
}

func (s *UserService) AuthUser(ctx context.Context, req *user.AuthUserRequest) (*model.UserEntity, error) {
	userEntity, err := s.CheckUser(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}

	if !util.CheckPasswordHash(req.GetPassword(), userEntity.Password) {
		logger.InfoL("Wrong password", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
		return nil, util.InvalidEmailOrPassword
	}

	return userEntity, nil
}

func (s *UserService) CheckUser(ctx context.Context, email string) (*model.UserEntity, error) {
	userEntity, err := s.userRepo.FindByEmail(email)
	if err != nil {
		logger.ErrorL("Unknown error", util.GetMetadata(ctx), logger.DT{"email": email})
		return nil, util.UnknownError(err)
	}

	if userEntity == nil {
		logger.InfoL("User not found by email", util.GetMetadata(ctx), logger.DT{"email": email})
		return nil, util.UserNotFound
	}

	return userEntity, nil
}

func (s *UserService) GetUser(ctx context.Context, req *user.GetUserRequest) (*model.UserEntity, error) {
	userEntity, err := s.userRepo.FindByEmail(req.GetEmail())
	if err != nil {
		return nil, util.UnknownError(err)
	}

	if userEntity == nil {
		logger.InfoL("User not found by email", util.GetMetadata(ctx), logger.DT{"email": req.GetEmail()})
		return nil, util.UserNotFound
	}

	return userEntity, nil
}

func (s *UserService) mapToEntityByUpdReq(req *user.UpdateUserRequest) *model.UserEntity {
	userEntity := &model.UserEntity{}

	if req.GetName() != "" {
		userEntity.Name = req.GetName()
	}

	if req.GetProfileImage() != "" {
		userEntity.ProfileImage = req.GetProfileImage()
	}

	if req.GetEmailVerifiedTime() > 0 {
		userEntity.EmailVerifiedTime = req.GetEmailVerifiedTime()
	}

	return userEntity
}
