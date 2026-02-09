package mcp

import (
	"encoding/json"
	"reflect"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// MCPQueryAllTransactionCategoriesResponse represents the response structure for querying transaction categories
type MCPQueryAllTransactionCategoriesResponse struct {
	IncomeCategories   []string `json:"incomeCategories" jsonschema_description:"List of income category names"`
	ExpenseCategories  []string `json:"expenseCategories" jsonschema_description:"List of expense category names"`
	TransferCategories []string `json:"transferCategories" jsonschema_description:"List of transfer category names"`
}

type mcpQueryAllTransactionCategoriesToolHandler struct{}

var MCPQueryAllTransactionCategoriesToolHandler = &mcpQueryAllTransactionCategoriesToolHandler{}

// Name returns the name of the MCP tool
func (h *mcpQueryAllTransactionCategoriesToolHandler) Name() string {
	return "query_all_transaction_categories"
}

// Description returns the description of the MCP tool
func (h *mcpQueryAllTransactionCategoriesToolHandler) Description() string {
	return "Query all transaction categories for the current user in ezBookkeeping."
}

// InputType returns the input type for the MCP tool request
func (h *mcpQueryAllTransactionCategoriesToolHandler) InputType() reflect.Type {
	return nil
}

// OutputType returns the output type for the MCP tool response
func (h *mcpQueryAllTransactionCategoriesToolHandler) OutputType() reflect.Type {
	return reflect.TypeOf(&MCPQueryAllTransactionCategoriesResponse{})
}

// Handle processes the MCP call tool request and returns the response
func (h *mcpQueryAllTransactionCategoriesToolHandler) Handle(c *core.WebContext, callToolReq *MCPCallToolRequest, user *models.User, currentConfig *settings.Config, services MCPAvailableServices) (any, []*MCPTextContent, error) {
	uid := user.Uid
	categories, err := services.GetTransactionCategoryService().GetAllCategoriesByUid(c, uid, 0)

	if err != nil {
		log.Errorf(c, "[query_all_transaction_categories.Handle] failed to get categories for user \"uid:%d\", because %s", uid, err.Error())
		return nil, nil, err
	}

	structuredResponse, response, err := h.createNewMCPQueryAllTransactionCategoriesResponse(c, categories)

	if err != nil {
		return nil, nil, err
	}

	return structuredResponse, response, nil
}

func (h *mcpQueryAllTransactionCategoriesToolHandler) createNewMCPQueryAllTransactionCategoriesResponse(c *core.WebContext, categories []*models.TransactionCategory) (any, []*MCPTextContent, error) {
	response := MCPQueryAllTransactionCategoriesResponse{
		IncomeCategories:   make([]string, 0),
		ExpenseCategories:  make([]string, 0),
		TransferCategories: make([]string, 0),
	}

	for i := 0; i < len(categories); i++ {
		category := categories[i]

		if category.Hidden {
			continue
		}

		if category.Type == models.CATEGORY_TYPE_INCOME {
			response.IncomeCategories = append(response.IncomeCategories, category.Name)
		} else if category.Type == models.CATEGORY_TYPE_EXPENSE {
			response.ExpenseCategories = append(response.ExpenseCategories, category.Name)
		} else if category.Type == models.CATEGORY_TYPE_TRANSFER {
			response.TransferCategories = append(response.TransferCategories, category.Name)
		}
	}

	content, err := json.Marshal(response)

	if err != nil {
		return nil, nil, err
	}

	return response, []*MCPTextContent{
		NewMCPTextContent(string(content)),
	}, nil
}
