package models

import (
	"strconv"
	"strings"
)

// TransactionSplit represents a split part of a transaction with its own category and amount
type TransactionSplit struct {
	SplitId         int64  `xorm:"PK NOT NULL"`
	Uid             int64  `xorm:"INDEX(IDX_transaction_split_uid_deleted_transaction_id) NOT NULL"`
	Deleted         bool   `xorm:"INDEX(IDX_transaction_split_uid_deleted_transaction_id) NOT NULL"`
	TransactionId   int64  `xorm:"INDEX(IDX_transaction_split_uid_deleted_transaction_id) NOT NULL"`
	CategoryId      int64  `xorm:"NOT NULL"`
	Amount          int64  `xorm:"NOT NULL"`
	SplitType       int32  `xorm:"NOT NULL DEFAULT 0"`
	TagIds          string `xorm:"VARCHAR(255) NOT NULL DEFAULT "`
	DisplayOrder    int32  `xorm:"NOT NULL"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

// GetTagIdSlice parses the comma-separated TagIds string into a slice of int64
func (s *TransactionSplit) GetTagIdSlice() []int64 {
	if s.TagIds == "" {
		return nil
	}
	parts := strings.Split(s.TagIds, ",")
	result := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err == nil && id > 0 {
			result = append(result, id)
		}
	}
	return result
}

// TagIdsFromSlice converts a slice of int64 tag IDs to a comma-separated string
func TagIdsFromSlice(tagIds []int64) string {
	if len(tagIds) == 0 {
		return ""
	}
	parts := make([]string, len(tagIds))
	for i, id := range tagIds {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ",")
}

// TransactionSplitCreateRequest represents the split part data in create/modify requests
type TransactionSplitCreateRequest struct {
	CategoryId int64   `json:"categoryId,string" binding:"required,min=1"`
	Amount     int64   `json:"amount" binding:"required,min=1"`
	SplitType  int32   `json:"splitType"`
	TagIds     []int64 `json:"tagIds"`
}

// TransactionSplitResponse represents the split part data in API responses
type TransactionSplitResponse struct {
	CategoryId int64   `json:"categoryId,string"`
	Amount     int64   `json:"amount"`
	SplitType  int32   `json:"splitType"`
	TagIds     []int64 `json:"tagIds"`
}
