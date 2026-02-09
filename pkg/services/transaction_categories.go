package services

import (
	"strings"
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// TransactionCategoryService represents transaction category service
type TransactionCategoryService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction category service singleton instance
var (
	TransactionCategories = &TransactionCategoryService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetTotalCategoryCountByUid returns total category count of user
func (s *TransactionCategoryService) GetTotalCategoryCountByUid(c core.Context, uid int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	count, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Count(&models.TransactionCategory{})

	return count, err
}

// GetAllCategoriesByUid returns all transaction category models of user
func (s *TransactionCategoryService) GetAllCategoriesByUid(c core.Context, uid int64, categoryType models.TransactionCategoryType) ([]*models.TransactionCategory, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	condition := "uid=? AND deleted=?"
	conditionParams := make([]any, 0, 8)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)

	if categoryType > 0 {
		condition = condition + " AND type=?"
		conditionParams = append(conditionParams, categoryType)
	}

	var categories []*models.TransactionCategory
	err := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...).OrderBy("type asc, display_order asc").Find(&categories)

	return categories, err
}

// GetCategoryByCategoryId returns a transaction category model according to transaction category id
func (s *TransactionCategoryService) GetCategoryByCategoryId(c core.Context, uid int64, categoryId int64) (*models.TransactionCategory, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if categoryId <= 0 {
		return nil, errs.ErrTransactionCategoryIdInvalid
	}

	category := &models.TransactionCategory{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(categoryId).Where("uid=? AND deleted=?", uid, false).Get(category)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrTransactionCategoryNotFound
	}

	return category, nil
}

// GetCategoriesByCategoryIds returns transaction category models according to transaction category ids
func (s *TransactionCategoryService) GetCategoriesByCategoryIds(c core.Context, uid int64, categoryIds []int64) (map[int64]*models.TransactionCategory, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if categoryIds == nil {
		return nil, errs.ErrTransactionCategoryIdInvalid
	}

	var categories []*models.TransactionCategory
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("category_id", categoryIds).Find(&categories)

	if err != nil {
		return nil, err
	}

	categoryMap := s.GetCategoryMapByList(categories)
	return categoryMap, err
}

// GetMaxDisplayOrder returns the max display order according to transaction category type
func (s *TransactionCategoryService) GetMaxDisplayOrder(c core.Context, uid int64, categoryType models.TransactionCategoryType) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	category := &models.TransactionCategory{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "display_order").Where("uid=? AND deleted=? AND type=?", uid, false, categoryType).OrderBy("display_order desc").Limit(1).Get(category)

	if err != nil {
		return 0, err
	}

	if has {
		return category.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// CreateCategory saves a new transaction category model to database
func (s *TransactionCategoryService) CreateCategory(c core.Context, category *models.TransactionCategory) error {
	if category.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	category.CategoryId = s.GenerateUuid(uuid.UUID_TYPE_CATEGORY)

	if category.CategoryId < 1 {
		return errs.ErrSystemIsBusy
	}

	category.ParentCategoryId = 0
	category.Deleted = false
	category.CreatedUnixTime = time.Now().Unix()
	category.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(category.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(category)
		return err
	})
}

// CreateCategories saves a few transaction category models to database
func (s *TransactionCategoryService) CreateCategories(c core.Context, uid int64, categories []*models.TransactionCategory) ([]*models.TransactionCategory, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	needCategoryUuidCount := uint16(len(categories))
	categoryUuids := s.GenerateUuids(uuid.UUID_TYPE_CATEGORY, needCategoryUuidCount)

	if len(categoryUuids) < int(needCategoryUuidCount) {
		return nil, errs.ErrSystemIsBusy
	}

	for i := 0; i < len(categories); i++ {
		categories[i].CategoryId = categoryUuids[i]
		categories[i].ParentCategoryId = 0
		categories[i].Deleted = false
		categories[i].CreatedUnixTime = time.Now().Unix()
		categories[i].UpdatedUnixTime = time.Now().Unix()
	}

	err := s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(categories); i++ {
			_, err := sess.Insert(categories[i])
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return categories, nil
}

// ModifyCategory saves an existed transaction category model to database
func (s *TransactionCategoryService) ModifyCategory(c core.Context, category *models.TransactionCategory) error {
	if category.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	category.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(category.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(category.CategoryId).Cols("name", "display_order", "icon", "color", "comment", "hidden", "updated_unix_time").Where("uid=? AND deleted=?", category.Uid, false).Update(category)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionCategoryNotFound
		}

		return nil
	})
}

// HideCategory updates hidden field of given transaction categories
func (s *TransactionCategoryService) HideCategory(c core.Context, uid int64, ids []int64, hidden bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionCategory{
		Hidden:          hidden,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("hidden", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).In("category_id", ids).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionCategoryNotFound
		}

		return nil
	})
}

// ModifyCategoryDisplayOrders updates display order of given transaction categories
func (s *TransactionCategoryService) ModifyCategoryDisplayOrders(c core.Context, uid int64, categories []*models.TransactionCategory) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(categories); i++ {
		categories[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(categories); i++ {
			category := categories[i]
			updatedRows, err := sess.ID(category.CategoryId).Cols("display_order", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(category)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrTransactionCategoryNotFound
			}
		}

		return nil
	})
}

// DeleteCategory deletes an existed transaction category from database
func (s *TransactionCategoryService) DeleteCategory(c core.Context, uid int64, categoryId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionCategory{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		exists, err := sess.Cols("uid", "deleted", "category_id").Where("uid=? AND deleted=? AND category_id=?", uid, false, categoryId).Limit(1).Exist(&models.Transaction{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionCategoryInUseCannotBeDeleted
		}

		exists, err = sess.Cols("uid", "deleted", "category_id", "template_type", "scheduled_frequency_type", "scheduled_end_time").Where("uid=? AND deleted=? AND (template_type=? OR (template_type=? AND scheduled_frequency_type<>? AND (scheduled_end_time IS NULL OR scheduled_end_time>=?)))", uid, false, models.TRANSACTION_TEMPLATE_TYPE_NORMAL, models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED, now).Where("category_id=?", categoryId).Limit(1).Exist(&models.TransactionTemplate{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionCategoryInUseCannotBeDeleted
		}

		deletedRows, err := sess.ID(categoryId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTransactionCategoryNotFound
		}

		return nil
	})
}

// DeleteAllCategories deletes all existed transaction categories from database
func (s *TransactionCategoryService) DeleteAllCategories(c core.Context, uid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionCategory{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		exists, err := sess.Cols("uid", "deleted", "category_id").Where("uid=? AND deleted=? AND category_id<>?", uid, false, 0).Limit(1).Exist(&models.Transaction{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionCategoryInUseCannotBeDeleted
		}

		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		}

		return nil
	})
}

// GetCategoryMapByList returns a transaction category map by a list
func (s *TransactionCategoryService) GetCategoryMapByList(categories []*models.TransactionCategory) map[int64]*models.TransactionCategory {
	categoryMap := make(map[int64]*models.TransactionCategory)

	for i := 0; i < len(categories); i++ {
		category := categories[i]
		categoryMap[category.CategoryId] = category
	}
	return categoryMap
}

// GetVisibleCategoryNameMapByList returns visible transaction category name maps by a list
func (s *TransactionCategoryService) GetVisibleCategoryNameMapByList(categories []*models.TransactionCategory) (expenseCategoryMap map[string]*models.TransactionCategory, incomeCategoryMap map[string]*models.TransactionCategory, transferCategoryMap map[string]*models.TransactionCategory) {
	expenseCategoryMap = make(map[string]*models.TransactionCategory)
	incomeCategoryMap = make(map[string]*models.TransactionCategory)
	transferCategoryMap = make(map[string]*models.TransactionCategory)

	for i := 0; i < len(categories); i++ {
		category := categories[i]

		if category.Hidden {
			continue
		}

		if category.Type == models.CATEGORY_TYPE_INCOME {
			incomeCategoryMap[category.Name] = category
		} else if category.Type == models.CATEGORY_TYPE_EXPENSE {
			expenseCategoryMap[category.Name] = category
		} else if category.Type == models.CATEGORY_TYPE_TRANSFER {
			transferCategoryMap[category.Name] = category
		}
	}

	return expenseCategoryMap, incomeCategoryMap, transferCategoryMap
}

// GetCategoryNames returns a list with transaction category names from transaction category models list
func (s *TransactionCategoryService) GetCategoryNames(categories []*models.TransactionCategory) []string {
	categoryNames := make([]string, len(categories))

	for i := 0; i < len(categories); i++ {
		categoryNames[i] = categories[i].Name
	}

	return categoryNames
}

// GetCategoryOrSubCategoryIds returns all category ids according to given category ids
func (s *TransactionCategoryService) GetCategoryOrSubCategoryIds(c core.Context, categoryIds string, uid int64) ([]int64, error) {
	if categoryIds == "" || categoryIds == "0" {
		return nil, nil
	}

	allCategoryIds, err := utils.StringArrayToInt64Array(strings.Split(categoryIds, ","))

	if err != nil {
		return nil, errs.Or(err, errs.ErrTransactionCategoryIdInvalid)
	}

	return allCategoryIds, nil
}

// GetCategoryOrSubCategoryIdsByCategoryName returns a list of transaction category ids according to given category name
func (s *TransactionCategoryService) GetCategoryOrSubCategoryIdsByCategoryName(categories []*models.TransactionCategory, categoryName string) []int64 {
	categoryIds := make([]int64, 0)

	for i := 0; i < len(categories); i++ {
		category := categories[i]
		if category.Name == categoryName {
			categoryIds = append(categoryIds, category.CategoryId)
		}
	}

	return categoryIds
}

// GetTransactionCategoryService returns the transaction category service instance
func GetTransactionCategoryService() *TransactionCategoryService {
	return TransactionCategories
}
