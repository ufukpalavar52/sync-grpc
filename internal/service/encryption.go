package service

import (
	"imapsync-grpc/internal/model"
	"imapsync-grpc/internal/repository"
	"imapsync-grpc/internal/util"

	"gorm.io/gorm"
)

type EncryptionService struct {
	encRepo *repository.EncryptionRepository
}

func NewEncryptionService(db *gorm.DB) *EncryptionService {
	return &EncryptionService{encRepo: repository.NewEncryptionRepository(db)}
}

func (s *EncryptionService) GetOrCreateFirstKey(version string) (*model.EncryptionKeyEntity, error) {
	enc, err := s.encRepo.FindByVersion(version)
	if err != nil {
		return nil, err
	}
	if enc != nil {
		return enc, nil
	}

	key, _, _ := util.GenerateRandomAESKey()

	create, err := s.encRepo.Create(key, version)
	if err != nil {
		return nil, err
	}
	return create, nil
}

func (s *EncryptionService) GetKeyByVersion(version string) (*model.EncryptionKeyEntity, error) {
	enc, err := s.encRepo.FindByVersion(version)
	if err != nil {
		return nil, util.UnknownError(err)
	}

	if enc != nil {
		return enc, nil
	}

	return nil, util.KeyVersionNotFound
}

func (s *EncryptionService) EncryptStr(str string, key []byte) (string, error) {
	enc, err := util.EncryptAESGCM(str, key)
	if err != nil {
		return "", util.UnknownError(err)
	}
	return enc, nil
}

func (s *EncryptionService) DecryptStr(str string, key []byte) (string, error) {
	enc, err := util.DecryptAESGCM(str, key)
	if err != nil {
		return "", util.UnknownError(err)
	}
	return enc, nil
}
