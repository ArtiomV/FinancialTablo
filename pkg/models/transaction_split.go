package models

// TransactionSplit represents a split part of a transaction with its own category and amount
type TransactionSplit struct {
	SplitId         int64 `xorm:"PK NOT NULL"`
	Uid             int64 `xorm:"INDEX(IDX_transaction_split_uid_deleted_transaction_id) NOT NULL"`
	Deleted         bool  `xorm:"INDEX(IDX_transaction_split_uid_deleted_transaction_id) NOT NULL"`
	TransactionId   int64 `xorm:"INDEX(IDX_transaction_split_uid_deleted_transaction_id) NOT NULL"`
	CategoryId      int64 `xorm:"NOT NULL"`
	Amount          int64 `xorm:"NOT NULL"`
	DisplayOrder    int32 `xorm:"NOT NULL"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

// TransactionSplitCreateRequest represents the split part data in create/modify requests
type TransactionSplitCreateRequest struct {
	CategoryId int64 `json:"categoryId,string" binding:"required,min=1"`
	Amount     int64 `json:"amount" binding:"required,min=1"`
}

// TransactionSplitResponse represents the split part data in API responses
type TransactionSplitResponse struct {
	CategoryId int64 `json:"categoryId,string"`
	Amount     int64 `json:"amount"`
}
