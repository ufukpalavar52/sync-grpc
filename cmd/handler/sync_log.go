package handler

import (
	"imapsync-grpc/internal/service"
	"imapsync-grpc/pkg/pb/sync_log"
	"log"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
)

type SyncLogService struct {
	sync_log.UnsafeSyncLogServiceServer
	validator   protovalidate.Validator
	tailService *service.SyncLogService
}

func NewSyncLogService() *SyncLogService {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}
	return &SyncLogService{validator: v, tailService: service.NewSyncLogService()}
}

func (s *SyncLogService) StreamLogs(req *sync_log.SyncLogRequest, stream grpc.ServerStreamingServer[sync_log.SyncLogChunk]) error {
	return s.tailService.Tail(req, stream)
}
