// transaction_splits.go handles splitting transactions across multiple categories.
package services

import (
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// TransactionSplitService represents transaction split service
type TransactionSplitService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction split service singleton instance
var (
	TransactionSplits = &TransactionSplitService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetSplitsByTransactionId returns split parts for a given transaction
func (s *TransactionSplitService) GetSplitsByTransactionId(c core.Context, uid int64, transactionId int64) ([]*models.TransactionSplit, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if transactionId <= 0 {
		return nil, errs.ErrTransactionIdInvalid
	}

	var splits []*models.TransactionSplit
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND transaction_id=?", uid, false, transactionId).OrderBy("display_order asc").Find(&splits)

	if err != nil {
		return nil, err
	}

	return splits, nil
}

// GetSplitsByTransactionIds returns split parts for multiple transactions (batch load)
func (s *TransactionSplitService) GetSplitsByTransactionIds(c core.Context, uid int64, transactionIds []int64) (map[int64][]*models.TransactionSplit, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if transactionIds == nil {
		return nil, errs.ErrTransactionIdInvalid
	}

	var splits []*models.TransactionSplit
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("transaction_id", transactionIds).OrderBy("display_order asc").Find(&splits)

	if err != nil {
		return nil, err
	}

	splitMap := make(map[int64][]*models.TransactionSplit)
	for _, split := range splits {
		splitMap[split.TransactionId] = append(splitMap[split.TransactionId], split)
	}

	return splitMap, nil
}

// CreateSplits creates split parts for a transaction (standalone with its own DB transaction)
func (s *TransactionSplitService) CreateSplits(c core.Context, uid int64, transactionId int64, splitRequests []models.TransactionSplitCreateRequest) error {
	if len(splitRequests) == 0 {
		return nil
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		return s.CreateSplitsInSession(sess, uid, transactionId, splitRequests)
	})
}

// ReplaceSplits replaces split parts for a transaction (standalone with its own DB transaction)
func (s *TransactionSplitService) ReplaceSplits(c core.Context, uid int64, transactionId int64, splitRequests []models.TransactionSplitCreateRequest) error {
	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		return s.ReplaceSplitsInSession(sess, uid, transactionId, splitRequests)
	})
}

// CreateSplitsInSession creates split parts within an existing database session
func (s *TransactionSplitService) CreateSplitsInSession(sess *xorm.Session, uid int64, transactionId int64, splitRequests []models.TransactionSplitCreateRequest) error {
	if len(splitRequests) == 0 {
		return nil
	}

	splitUuids := s.GenerateUuids(uuid.UUID_TYPE_SPLIT, uint16(len(splitRequests)))

	if len(splitUuids) < len(splitRequests) {
		return errs.ErrSystemIsBusy
	}

	now := time.Now().Unix()

	for i, req := range splitRequests {
		split := &models.TransactionSplit{
			SplitId:         splitUuids[i],
			Uid:             uid,
			Deleted:         false,
			TransactionId:   transactionId,
			CategoryId:      req.CategoryId,
			Amount:          req.Amount,
			DisplayOrder:    int32(i),
			CreatedUnixTime: now,
			UpdatedUnixTime: now,
		}

		_, err := sess.Insert(split)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteSplitsInSession soft-deletes all splits for a transaction within an existing session
func (s *TransactionSplitService) DeleteSplitsInSession(sess *xorm.Session, uid int64, transactionId int64) error {
	now := time.Now().Unix()

	updateModel := &models.TransactionSplit{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	_, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", uid, false, transactionId).Update(updateModel)
	return err
}

// ReplaceSplitsInSession deletes old splits and creates new ones within a session
func (s *TransactionSplitService) ReplaceSplitsInSession(sess *xorm.Session, uid int64, transactionId int64, splitRequests []models.TransactionSplitCreateRequest) error {
	// Delete old splits
	err := s.DeleteSplitsInSession(sess, uid, transactionId)
	if err != nil {
		return err
	}

	// Create new splits (if any)
	return s.CreateSplitsInSession(sess, uid, transactionId, splitRequests)
}
