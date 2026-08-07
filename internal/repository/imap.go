package repository

import (
	"errors"
	"imapsync-user/internal/model"
	"imapsync-user/internal/util"
	"imapsync-user/pkg/pb/sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ImapRepository struct {
	db   *gorm.DB
	base *BaseRepository
}

func NewImapRepository(db *gorm.DB) *ImapRepository {
	return &ImapRepository{db: db, base: NewBaseRepository(db)}
}

func (r *ImapRepository) CreateSync(request *sync.ImapSyncRequest, encVersion string) (*model.ImapSyncEntity, error) {
	entity := model.ImapSyncEntity{
		TransactionId:      request.GetTransactionId(),
		UserID:             request.GetUserId(),
		SourceUser:         request.GetSourceUser(),
		SourceHost:         request.GetSourceHost(),
		SourcePassword:     request.GetSourcePassword(),
		SourceAuthUser:     request.GetSourceAuthUser(),
		SourceSSL:          request.GetSourceSSL(),
		SourceTenantID:     request.GetSourceTenantId(),
		SourceClientID:     request.GetSourceClientId(),
		SourceClientSecret: request.GetSourceClientSecret(),
		SourcePort:         request.GetSourcePort(),
		DestUser:           request.GetDestUser(),
		DestAuthUser:       request.GetDestAuthUser(),
		DestPassword:       request.GetDestPassword(),
		DestSSL:            request.GetDestSSL(),
		DestTenantID:       request.GetDestTenantId(),
		DestClientID:       request.GetDestClientId(),
		DestClientSecret:   request.GetDestClientSecret(),
		DestPort:           request.GetDestPort(),
		SkipHeader:         request.GetSkipHeader(),
		DestHost:           request.GetDestHost(),
		EncryptionVersion:  encVersion,
		Status:             util.ImapPending,
		ExcludeFolders:     request.ExcludeFolders,
	}

	tx := r.db.Create(&entity)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &entity, nil
}

func (r *ImapRepository) GetSyncByTransactionId(transactionId string) (*model.ImapSyncEntity, error) {
	var entity model.ImapSyncEntity
	tx := r.db.Where(&model.ImapSyncEntity{TransactionId: transactionId}).First(&entity)

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &entity, nil
}

func (r *ImapRepository) UpdateById(id int64, update any) (*model.ImapSyncEntity, error) {
	var entity model.ImapSyncEntity
	tx := r.db.Where("id = ?", id).Updates(update)

	if tx.Error != nil {
		return nil, tx.Error
	}

	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	r.db.Where("id = ?", id).First(&entity)

	return &entity, nil
}

func (r *ImapRepository) FindAllBy(search any, p model.Pagination) (*model.PaginationResult[[]model.ImapSyncEntity], error) {
	db := r.base.ApplySmartFilters(search, true)
	var entities []model.ImapSyncEntity
	var total int64
	checkTotal := false
	if p.Limit > 0 {
		db = db.Limit(p.Limit)
		checkTotal = true
	}

	if p.Offset > 0 {
		db = db.Offset(p.Offset)
		checkTotal = true
	}
	tx := db.Order(clause.OrderByColumn{
		Column: clause.Column{Name: "id"},
		Desc:   true,
	}).Find(&entities)
	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, tx.Error
	}

	if checkTotal {
		newDb := r.base.ApplySmartFilters(search, true)
		tx = newDb.Model(&model.ImapSyncEntity{}).Count(&total)
	}

	return model.NewPaginationResult[[]model.ImapSyncEntity](total, entities), nil
}

func (r *ImapRepository) GetCountByStatusStats(search any) ([]model.SyncCountGroupByStatus, error) {
	query := r.db.Model(&model.ImapSyncEntity{})
	if search != nil {
		query = query.Where(search)
	}

	var results []model.SyncCountGroupByStatus

	err := query.
		Select("status, count(id) as count, AVG(finish_time) as avg_time").
		Group("status").Scan(&results).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return results, nil
	}

	if err != nil {
		return nil, err
	}

	return results, nil
}
