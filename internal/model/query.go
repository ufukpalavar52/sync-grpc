package model

import "imapsync-user/pkg/pb/sync"

type SearchSync struct {
	UserId        int64  `name:"userId"`
	TransactionId string `name:"transactionId"`
	SourceHost    string `name:"sourceHost" filter:"like"`
	SourceUser    string `name:"sourceUser" filter:"like"`
	DestHost      string `name:"destHost" filter:"like"`
	DestUser      string `name:"destUser" filter:"like"`
	Status        string `name:"status"`
}

type SyncCountGroupByStatus struct {
	Count   int64   `name:"count"`
	Status  string  `name:"status"`
	AvgTime float64 `name:"avgTime" gorm:"column:avg_time"`
}

func NewSearchSyncFromReq(request *sync.ListSyncRequest) *SearchSync {
	return &SearchSync{
		UserId:        request.GetUserId(),
		TransactionId: request.GetTransactionId(),
		SourceHost:    request.GetSourceHost(),
		SourceUser:    request.GetSourceUser(),
		DestHost:      request.GetDestHost(),
		DestUser:      request.GetDestUser(),
		Status:        request.GetStatus(),
	}
}

type Pagination struct {
	Limit  int
	Offset int
}

type PaginationResult[T any] struct {
	total int64
	data  T
}

func (p *PaginationResult[T]) GetData() T {
	return p.data
}

func (p *PaginationResult[T]) GetTotal() int64 {
	return p.total
}

func NewPaginationResult[T any](total int64, data T) *PaginationResult[T] {
	return &PaginationResult[T]{total, data}
}
