package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func TestGetCategoryMapByList_EmptyList(t *testing.T) {
	categories := make([]*models.TransactionCategory, 0)
	actualCategoryMap := TransactionCategories.GetCategoryMapByList(categories)

	assert.NotNil(t, actualCategoryMap)
	assert.Equal(t, 0, len(actualCategoryMap))
}

func TestGetCategoryMapByList_MultipleCategories(t *testing.T) {
	categories := []*models.TransactionCategory{
		{
			CategoryId: 1001,
			Name:       "Category Name",
			Type:       models.CATEGORY_TYPE_EXPENSE,
			Hidden:     false,
		},
		{
			CategoryId: 1002,
			Name:       "Category Name2",
			Type:       models.CATEGORY_TYPE_INCOME,
			Hidden:     false,
		},
		{
			CategoryId: 1003,
			Name:       "Category Name3",
			Type:       models.CATEGORY_TYPE_TRANSFER,
			Hidden:     true,
		},
	}
	actualCategoryMap := TransactionCategories.GetCategoryMapByList(categories)

	assert.Equal(t, 3, len(actualCategoryMap))
	assert.Contains(t, actualCategoryMap, int64(1001))
	assert.Contains(t, actualCategoryMap, int64(1002))
	assert.Contains(t, actualCategoryMap, int64(1003))
	assert.Equal(t, "Category Name", actualCategoryMap[1001].Name)
	assert.Equal(t, "Category Name2", actualCategoryMap[1002].Name)
	assert.Equal(t, "Category Name3", actualCategoryMap[1003].Name)
}

func TestGetVisibleCategoryNameMapByList_EmptyList(t *testing.T) {
	categories := make([]*models.TransactionCategory, 0)
	expenseCategoryMap, incomeCategoryMap, transferCategoryMap := TransactionCategories.GetVisibleCategoryNameMapByList(categories)

	assert.NotNil(t, expenseCategoryMap)
	assert.NotNil(t, incomeCategoryMap)
	assert.NotNil(t, transferCategoryMap)
	assert.Equal(t, 0, len(expenseCategoryMap))
	assert.Equal(t, 0, len(incomeCategoryMap))
	assert.Equal(t, 0, len(transferCategoryMap))
}

func TestGetVisibleCategoryNameMapByList_WithHiddenCategories(t *testing.T) {
	categories := []*models.TransactionCategory{
		{
			CategoryId: 1001,
			Name:       "Category Name",
			Type:       models.CATEGORY_TYPE_EXPENSE,
			Hidden:     true,
		},
		{
			CategoryId: 1002,
			Name:       "Category Name2",
			Type:       models.CATEGORY_TYPE_EXPENSE,
			Hidden:     false,
		},
	}
	expenseCategoryMap, incomeCategoryMap, transferCategoryMap := TransactionCategories.GetVisibleCategoryNameMapByList(categories)

	assert.Equal(t, 1, len(expenseCategoryMap))
	assert.Contains(t, expenseCategoryMap, "Category Name2")
	assert.NotContains(t, expenseCategoryMap, "Category Name")
	assert.Equal(t, 0, len(incomeCategoryMap))
	assert.Equal(t, 0, len(transferCategoryMap))
}

func TestGetVisibleCategoryNameMapByList_AllTypes(t *testing.T) {
	categories := []*models.TransactionCategory{
		{
			CategoryId: 1001,
			Name:       "Category Name",
			Type:       models.CATEGORY_TYPE_EXPENSE,
			Hidden:     false,
		},
		{
			CategoryId: 1002,
			Name:       "Category Name2",
			Type:       models.CATEGORY_TYPE_INCOME,
			Hidden:     false,
		},
		{
			CategoryId: 1003,
			Name:       "Category Name3",
			Type:       models.CATEGORY_TYPE_TRANSFER,
			Hidden:     false,
		},
	}
	expenseCategoryMap, incomeCategoryMap, transferCategoryMap := TransactionCategories.GetVisibleCategoryNameMapByList(categories)

	assert.Equal(t, 1, len(expenseCategoryMap))
	assert.Contains(t, expenseCategoryMap, "Category Name")

	assert.Equal(t, 1, len(incomeCategoryMap))
	assert.Contains(t, incomeCategoryMap, "Category Name2")

	assert.Equal(t, 1, len(transferCategoryMap))
	assert.Contains(t, transferCategoryMap, "Category Name3")
}

func TestGetCategoryNames_EmptyList(t *testing.T) {
	categories := make([]*models.TransactionCategory, 0)
	actualNames := TransactionCategories.GetCategoryNames(categories)

	assert.NotNil(t, actualNames)
	assert.Equal(t, 0, len(actualNames))
}

func TestGetCategoryNames_MultipleCategories(t *testing.T) {
	categories := []*models.TransactionCategory{
		{
			CategoryId: 1001,
			Name:       "Category Name",
		},
		{
			CategoryId: 1002,
			Name:       "Category Name2",
		},
		{
			CategoryId: 1003,
			Name:       "Category Name3",
		},
	}
	actualNames := TransactionCategories.GetCategoryNames(categories)

	assert.Equal(t, 3, len(actualNames))
	assert.Equal(t, "Category Name", actualNames[0])
	assert.Equal(t, "Category Name2", actualNames[1])
	assert.Equal(t, "Category Name3", actualNames[2])
}

func TestGetCategoryOrSubCategoryIdsByCategoryName_EmptyList(t *testing.T) {
	categories := make([]*models.TransactionCategory, 0)
	actualIds := TransactionCategories.GetCategoryOrSubCategoryIdsByCategoryName(categories, "Category Name")

	assert.NotNil(t, actualIds)
	assert.Equal(t, 0, len(actualIds))
}

func TestGetCategoryOrSubCategoryIdsByCategoryName_NotExistName(t *testing.T) {
	categories := []*models.TransactionCategory{
		{
			CategoryId: 1001,
			Name:       "Category Name",
		},
	}
	actualIds := TransactionCategories.GetCategoryOrSubCategoryIdsByCategoryName(categories, "Non-existent Category")

	assert.NotNil(t, actualIds)
	assert.Equal(t, 0, len(actualIds))
}

func TestGetCategoryOrSubCategoryIdsByCategoryName_MatchByName(t *testing.T) {
	categories := []*models.TransactionCategory{
		{
			CategoryId: 1001,
			Name:       "Category Name",
		},
		{
			CategoryId: 1002,
			Name:       "Category Name",
		},
		{
			CategoryId: 1003,
			Name:       "Category Name2",
		},
	}
	actualIds := TransactionCategories.GetCategoryOrSubCategoryIdsByCategoryName(categories, "Category Name")

	assert.Equal(t, 2, len(actualIds))
	assert.Contains(t, actualIds, int64(1001))
	assert.Contains(t, actualIds, int64(1002))
}
