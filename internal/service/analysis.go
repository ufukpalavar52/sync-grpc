package service

import (
	"imapsync-user/internal/model"
	"imapsync-user/internal/repository"
	"imapsync-user/internal/util"
	"imapsync-user/pkg/pb/analysis"

	"gorm.io/gorm"
)

type AnalysisService struct {
	imapRepo *repository.ImapRepository
}

func NewAnalysisService(db *gorm.DB) *AnalysisService {
	return &AnalysisService{
		imapRepo: repository.NewImapRepository(db),
	}
}

func (s *AnalysisService) GetSyncCountStats(req *analysis.SyncsStatsRequest) (*analysis.SyncsStatsResponseByStatus, error) {
	var search any = nil
	if req.GetUserId() > 0 {
		search = model.ImapSyncEntity{UserID: req.GetUserId()}
	}

	return s.transformSyncCountStats(s.imapRepo.GetCountByStatusStats(search))
}

func (s *AnalysisService) transformSyncCountStats(stats []model.SyncCountGroupByStatus, err error) (*analysis.SyncsStatsResponseByStatus, error) {
	if err != nil {
		return nil, util.UnknownError(err)
	}

	response := &analysis.SyncsStatsResponseByStatus{}
	totalAvgTime := float64(0)
	totalCount := int64(0)
	for _, stat := range stats {
		totalCount += stat.Count
		totalAvgTime += stat.AvgTime * float64(stat.Count)
		response.Stats = append(response.Stats, &analysis.SyncsStatusStats{
			Status:  stat.Status,
			Count:   stat.Count,
			AvgTime: stat.AvgTime,
		})
	}

	if totalCount > 0 {
		totalAvgTime = totalAvgTime / float64(totalCount)
	}

	response.Stats = append(response.Stats, &analysis.SyncsStatusStats{
		Status:  "Total",
		Count:   totalCount,
		AvgTime: totalAvgTime,
	})

	return response, nil
}
