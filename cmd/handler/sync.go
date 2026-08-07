package handler

import (
	"context"
	"imapsync-grpc/internal/model"
	"imapsync-grpc/internal/service"
	"imapsync-grpc/internal/util"
	"imapsync-grpc/logger"
	"imapsync-grpc/pkg/pb/sync"
	"log"
	"time"

	"buf.build/go/protovalidate"
	"gorm.io/gorm"
)

type SyncServer struct {
	sync.UnimplementedSyncServiceServer
	validator   protovalidate.Validator
	imapService *service.ImapService
}

func NewSyncServer(db *gorm.DB) *SyncServer {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}
	return &SyncServer{validator: v, imapService: service.NewImapService(db)}
}

func (s *SyncServer) CreateSync(ctx context.Context, req *sync.ImapSyncRequest) (*sync.ImapSyncResponse, error) {
	md := util.GetMetadata(ctx)
	defer logger.InfoL("CreateSync request was finished.", md, logger.DT{"transactionId": req.GetTransactionId(), "userId": req.GetUserId()})
	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}
	return s.transformSyncResponse(s.imapService.CreateSync(req))
}

func (s *SyncServer) GetSync(ctx context.Context, req *sync.GetSyncRequest) (*sync.ImapSyncResponse, error) {
	md := util.GetMetadata(ctx)
	defer logger.InfoL("GetSync request was finished.", md, logger.DT{"transactionId": req.GetTransactionId()})

	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}

	return s.transformSyncResponse(s.imapService.GetSync(req))
}

func (s *SyncServer) GetSyncFullDetail(ctx context.Context, req *sync.GetSyncRequest) (*sync.ImapSyncFullResponse, error) {
	md := util.GetMetadata(ctx)
	defer logger.InfoL("GetSync request was finished.", md, logger.DT{"transactionId": req.GetTransactionId()})

	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}

	return s.transformSyncFullResponse(s.imapService.GetSync(req))
}

func (s *SyncServer) BulkCreateSync(ctx context.Context, req *sync.BulkCreateSyncRequest) (*sync.BulkCreateSyncResponse, error) {
	md := util.GetMetadata(ctx)
	defer logger.InfoL("BulkCreateSync request was finished.", md)
	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}

	res, errs := s.imapService.CreateBulk(req)
	var syncResponses []*sync.ImapSyncResponse
	for _, r := range res {
		transform, _ := s.transformSyncResponse(r, nil)
		syncResponses = append(syncResponses, transform)
	}

	return &sync.BulkCreateSyncResponse{
		Syncs:  syncResponses,
		Errors: errs,
	}, nil
}

func (s *SyncServer) ListSync(ctx context.Context, req *sync.ListSyncRequest) (*sync.ListSyncResponse, error) {
	md := util.GetMetadata(ctx)
	defer logger.InfoL("GetSync request was finished.", md, logger.DT{"transactionId": req.GetTransactionId(), "userId": req.GetUserId()})

	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}

	res, err := s.imapService.ListSyncs(req)

	if err != nil {
		return nil, err
	}

	return &sync.ListSyncResponse{
		Syncs: s.transformSyncListResponse(res.GetData()),
		Total: res.GetTotal(),
	}, nil
}

func (s *SyncServer) UpdateStatus(ctx context.Context, req *sync.UpdateStatusRequest) (*sync.ImapSyncResponse, error) {
	md := util.GetMetadata(ctx)
	defer logger.InfoL("UpdateStatus request was finished.", md, logger.DT{"transactionId": req.GetTransactionId(), "status": req.GetStatus()})
	if err := s.validator.Validate(req); err != nil {
		return nil, util.InvalidArgument(err)
	}

	return s.transformSyncResponse(s.imapService.UpdateStatus(req))
}

func (s *SyncServer) transformSyncResponse(entity *model.ImapSyncEntity, err error) (*sync.ImapSyncResponse, error) {
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, util.SyncNotFound
	}

	return &sync.ImapSyncResponse{
		Id:            entity.ID,
		TransactionId: entity.TransactionId,
		UserId:        entity.UserID,
		SourceUser:    entity.SourceUser,
		SourceHost:    entity.SourceHost,
		DestUser:      entity.DestUser,
		DestHost:      entity.DestHost,
		Status:        entity.Status,
		CreatedAt:     entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     entity.UpdatedAt.Format(time.RFC3339),
		FinishedTime:  entity.FinishTime,
	}, nil
}

func (s *SyncServer) transformSyncFullResponse(entity *model.ImapSyncEntity, err error) (*sync.ImapSyncFullResponse, error) {
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, util.SyncNotFound
	}

	return &sync.ImapSyncFullResponse{
		TransactionId:  entity.TransactionId,
		UserId:         entity.UserID,
		SourceUser:     entity.SourceUser,
		SourceHost:     entity.SourceHost,
		DestUser:       entity.DestUser,
		DestHost:       entity.DestHost,
		Status:         entity.Status,
		SourcePassword: entity.SourcePassword,
		DestPassword:   entity.DestPassword,

		SourceSSL:  entity.SourceSSL,
		SourcePort: entity.SourcePort,

		DestSSL:            entity.DestSSL,
		DestPort:           entity.DestPort,
		SourceAuthUser:     entity.SourceAuthUser,
		SourceTenantId:     entity.SourceTenantID,
		SourceClientId:     entity.SourceClientID,
		SourceClientSecret: entity.SourceClientSecret,
		DestAuthUser:       entity.DestAuthUser,
		DestTenantId:       entity.DestTenantID,
		DestClientId:       entity.DestClientID,
		DestClientSecret:   entity.DestClientSecret,
		SkipHeader:         entity.SkipHeader,
		ExcludeFolders:     entity.ExcludeFolders,
	}, nil
}

func (s *SyncServer) transformSyncListResponse(entityList []model.ImapSyncEntity) []*sync.ImapSyncResponse {
	var imapSyncList []*sync.ImapSyncResponse
	for _, e := range entityList {
		res, _ := s.transformSyncResponse(&e, nil)
		imapSyncList = append(imapSyncList, res)
	}

	return imapSyncList
}
