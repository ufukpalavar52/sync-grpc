package handler

import (
	"context"
	"imapsync-user/config"
	"imapsync-user/internal/global"
	"imapsync-user/internal/service"
	"imapsync-user/internal/util"
	"imapsync-user/logger"
	"imapsync-user/pkg/pb/encrypt"
	"log"

	"buf.build/go/protovalidate"
	"gorm.io/gorm"
)

type EncryptServer struct {
	encrypt.UnimplementedEncryptServiceServer
	validator  protovalidate.Validator
	encService *service.EncryptionService
}

func NewEncryptServer(db *gorm.DB) *EncryptServer {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}
	return &EncryptServer{validator: v, encService: service.NewEncryptionService(db)}
}

func (s *EncryptServer) Encrypt(ctx context.Context, req *encrypt.EncryptRequest) (*encrypt.EncryptResponse, error) {
	defer logger.InfoL("Encrypt request was finished.", util.GetMetadata(ctx))
	key, version, err := s.GetKeyAndVersion(req)
	if err != nil {
		return nil, err
	}

	str, err := s.encService.EncryptStr(req.GetData(), key)
	if err != nil {
		return nil, err
	}

	return &encrypt.EncryptResponse{
		Data:    str,
		Version: version,
	}, nil
}

func (s *EncryptServer) Decrypt(ctx context.Context, req *encrypt.EncryptRequest) (*encrypt.EncryptResponse, error) {
	defer logger.InfoL("Decrypt request was finished.", util.GetMetadata(ctx))
	key, version, err := s.GetKeyAndVersion(req)
	if err != nil {
		return nil, err
	}
	str, err := s.encService.DecryptStr(req.GetData(), key)
	if err != nil {
		return nil, err
	}

	return &encrypt.EncryptResponse{
		Data:    str,
		Version: version,
	}, nil
}

func (s *EncryptServer) GetKeyAndVersion(req *encrypt.EncryptRequest) ([]byte, string, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, "", util.InvalidArgument(err)
	}
	version := req.GetVersion()
	var key []byte
	if version == "" {
		version = config.EncryptionKeyVersion
		key = global.EncKey.GetByteKey()
	}

	if key == nil {
		enc, err := s.encService.GetKeyByVersion(version)
		if err != nil {
			return nil, "", util.UnknownError(err)
		}
		key = enc.GetByteKey()
	}

	return key, version, nil
}
