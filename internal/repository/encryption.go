package repository

import (
	"errors"
	"imapsync-grpc/internal/model"

	"gorm.io/gorm"
)

type EncryptionRepository struct {
	db *gorm.DB
}

func NewEncryptionRepository(db *gorm.DB) *EncryptionRepository {
	return &EncryptionRepository{db: db}
}

func (r *EncryptionRepository) Create(plaintext string, version string) (*model.EncryptionKeyEntity, error) {
	key := model.EncryptionKeyEntity{Key: plaintext, Version: version}
	tx := r.db.Create(&key)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &key, nil
}

func (r *EncryptionRepository) FindByVersion(version string) (*model.EncryptionKeyEntity, error) {
	var key model.EncryptionKeyEntity
	tx := r.db.Where(&model.EncryptionKeyEntity{Version: version}).First(&key)

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &key, nil
}
