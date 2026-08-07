package model

import (
	"imapsync-grpc/internal/util"
	"time"
)

type UserEntity struct {
	ID                int64 `gorm:"primary_key"`
	Name              string
	Email             string
	Password          string
	ProfileImage      string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	EmailVerifiedTime int64
}

func (*UserEntity) TableName() string {
	return "users"
}

type ImapSyncEntity struct {
	ID                 int64  `gorm:"primary_key"`
	TransactionId      string `gorm:"uniqueIndex"`
	SourceHost         string
	SourceUser         string
	SourceAuthUser     string
	SourcePassword     string
	SourceSSL          bool
	SourceTenantID     string
	SourceClientID     string
	SourceClientSecret string
	SourcePort         int32
	DestHost           string
	DestUser           string
	DestAuthUser       string
	DestPassword       string
	DestSSL            bool
	DestTenantID       string
	DestClientID       string
	DestClientSecret   string
	DestPort           int32
	SkipHeader         []string `gorm:"type:jsonb;serializer:json"`
	Status             string
	UserID             int64
	EncryptionVersion  string
	ExcludeFolders     []string `gorm:"type:jsonb;serializer:json"`
	FinishTime         float64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (*ImapSyncEntity) TableName() string {
	return "imap_syncs"
}

type EncryptionKeyEntity struct {
	ID        int64 `gorm:"primary_key"`
	Key       string
	Version   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (*EncryptionKeyEntity) TableName() string {
	return "encryption_keys"
}

func (e *EncryptionKeyEntity) GetByteKey() []byte {
	return util.Base64Decode(e.Key)
}
