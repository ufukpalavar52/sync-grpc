package service

import (
	"errors"
	"imapsync-user/internal/global"
	"imapsync-user/internal/model"
	"imapsync-user/internal/repository"
	"imapsync-user/internal/util"
	"imapsync-user/pkg/pb/sync"
	"time"

	"gorm.io/gorm"
)

type ImapService struct {
	imapRepo   *repository.ImapRepository
	userRepo   *repository.UserRepository
	encService *EncryptionService
}

func NewImapService(db *gorm.DB) *ImapService {
	return &ImapService{
		imapRepo:   repository.NewImapRepository(db),
		userRepo:   repository.NewUserRepository(db),
		encService: NewEncryptionService(db),
	}
}

func (s *ImapService) CreateSync(req *sync.ImapSyncRequest) (*model.ImapSyncEntity, error) {
	userEntity, err := s.userRepo.FindById(req.GetUserId())

	if err != nil {
		return nil, util.UnknownError(err)
	}

	if userEntity == nil {
		return nil, util.UserNotFound
	}

	imapSync, err := s.imapRepo.GetSyncByTransactionId(req.GetTransactionId())
	if imapSync != nil {
		return nil, util.SyncAlreadyExists
	}

	if err != nil {
		return nil, util.UnknownError(err)
	}

	if req.IsEncrypted == nil || *req.IsEncrypted == false {
		err = s.EncRequest(req)
		if err != nil {
			return nil, util.UnknownError(err)
		}
	}

	create, err := s.imapRepo.CreateSync(req, global.EncKey.Version)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, util.SyncAlreadyExists
		}
		return nil, util.UnknownError(err)
	}

	return create, nil

}

func (s *ImapService) CreateBulk(req *sync.BulkCreateSyncRequest) ([]*model.ImapSyncEntity, []*sync.ErrorResponse) {
	var errors []*sync.ErrorResponse
	var syncResponses []*model.ImapSyncEntity
	for _, r := range req.Syncs {
		res, err := s.CreateSync(r)
		if err != nil {
			errors = append(errors, &sync.ErrorResponse{
				TransactionId: r.TransactionId,
				Message:       err.Error(),
			})
			continue
		}
		syncResponses = append(syncResponses, res)
	}

	return syncResponses, errors
}

func (s *ImapService) GetSync(req *sync.GetSyncRequest) (*model.ImapSyncEntity, error) {
	syncRes, err := s.imapRepo.GetSyncByTransactionId(req.GetTransactionId())
	if err != nil {
		return nil, util.UnknownError(err)
	}

	if syncRes == nil {
		return nil, util.SyncNotFound
	}

	return syncRes, nil
}

func (s *ImapService) ListSyncs(req *sync.ListSyncRequest) (*model.PaginationResult[[]model.ImapSyncEntity], error) {
	search := model.NewSearchSyncFromReq(req)
	return s.imapRepo.FindAllBy(search, model.Pagination{Limit: int(req.Limit), Offset: int(req.Offset)})
}

func (s *ImapService) UpdateStatus(req *sync.UpdateStatusRequest) (*model.ImapSyncEntity, error) {
	syncRes, err := s.imapRepo.GetSyncByTransactionId(req.GetTransactionId())
	if err != nil {
		return nil, util.UnknownError(err)
	}

	if syncRes == nil {
		return nil, util.SyncNotFound
	}
	updateData := &model.ImapSyncEntity{Status: req.Status}
	if req.GetStatus() != util.ImapInProgress && req.Status != util.ImapPending {
		updateData.FinishTime = time.Since(syncRes.CreatedAt).Seconds()
	}

	imapEntity, err := s.imapRepo.UpdateById(syncRes.ID, updateData)
	if err != nil {
		return nil, util.UnknownError(err)
	}

	return imapEntity, nil
}

func (s *ImapService) EncRequest(req *sync.ImapSyncRequest) error {
	var err error

	req.SourcePassword, err = s.encService.EncryptStr(req.GetSourcePassword(), global.EncKey.GetByteKey())
	if err != nil {
		return err
	}

	req.DestPassword, err = s.encService.EncryptStr(req.GetDestPassword(), global.EncKey.GetByteKey())

	if err != nil {
		return err
	}

	if req.GetDestClientSecret() != "" {
		secret, errE := s.encService.EncryptStr(req.GetDestClientSecret(), global.EncKey.GetByteKey())
		if errE != nil {
			return errE
		}
		req.DestClientSecret = &secret
	}

	if req.GetSourceClientSecret() != "" {
		secret, errE := s.encService.EncryptStr(req.GetSourceClientSecret(), global.EncKey.GetByteKey())
		if errE != nil {
			return errE
		}
		req.SourceClientSecret = &secret
	}

	return nil
}
