package handler

import (
	"context"
	"imapsync-user/internal/service"
	"imapsync-user/pkg/pb/analysis"
	"log"

	"buf.build/go/protovalidate"
	"gorm.io/gorm"
)

type AnalysisServer struct {
	analysis.UnimplementedAnalysisServiceServer
	validator protovalidate.Validator
	anService *service.AnalysisService
}

func NewAnalysisServer(db *gorm.DB) *AnalysisServer {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}
	return &AnalysisServer{validator: v, anService: service.NewAnalysisService(db)}
}

func (s *AnalysisServer) GetSyncCountStats(ctx context.Context, req *analysis.SyncsStatsRequest) (*analysis.SyncsStatsResponseByStatus, error) {
	return s.anService.GetSyncCountStats(req)
}
