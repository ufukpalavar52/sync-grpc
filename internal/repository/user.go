package repository

import (
	"errors"
	"imapsync-user/internal/model"
	"imapsync-user/pkg/pb/user"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) CreateByRequest(request *user.CreateUserRequest) (*model.UserEntity, error) {
	userModel := model.UserEntity{
		Email:             request.GetEmail(),
		Password:          request.GetPassword(),
		Name:              request.GetName(),
		ProfileImage:      request.GetProfileImage(),
		EmailVerifiedTime: request.GetEmailVerifiedTime(),
	}
	tx := r.db.Create(&userModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &userModel, nil
}

func (r *UserRepository) UpdateByRequest(request *user.UpdateUserRequest) (*model.UserEntity, error) {
	userModel := model.UserEntity{
		Name:         request.GetName(),
		ProfileImage: request.GetProfileImage(),
	}
	tx := r.db.Where(model.UserEntity{Email: request.GetEmail()}).Updates(&userModel)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return r.FindByEmail(request.GetEmail())
}

func (r *UserRepository) UpdateByModel(id int64, userModel *model.UserEntity) (*model.UserEntity, error) {
	tx := r.db.Where(model.UserEntity{ID: id}).Updates(userModel)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return r.FindById(id)
}

func (r *UserRepository) UpdatePassByRequest(request *user.UpdateUserPasswordRequest) (*model.UserEntity, error) {
	userModel := model.UserEntity{
		Password: request.GetNewPassword(),
	}
	tx := r.db.Where(model.UserEntity{Email: request.GetEmail()}).Updates(&userModel)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return r.FindByEmail(request.GetEmail())
}

func (r *UserRepository) FindByEmail(email string) (*model.UserEntity, error) {
	return r.FindByFirst(model.UserEntity{Email: email})
}

func (r *UserRepository) FindById(id int64) (*model.UserEntity, error) {
	return r.FindByFirst(model.UserEntity{ID: id})
}

func (r *UserRepository) FindByFirst(query any) (*model.UserEntity, error) {
	var userModel model.UserEntity
	tx := r.db.Where(query).First(&userModel)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &userModel, nil
}
