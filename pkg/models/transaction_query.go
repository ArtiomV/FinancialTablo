package models

// TransactionQueryParams represents common query parameters for transaction listing and counting
type TransactionQueryParams struct {
	Uid                int64
	MaxTransactionTime int64
	MinTransactionTime int64
	TransactionType    TransactionType
	CategoryIds        []int64
	AccountIds         []int64
	TagFilters         []*TransactionTagFilter
	NoTags             bool
	AmountFilter       string
	Keyword            string
	CounterpartyId     int64
	Page               int32
	Count              int32
	NeedOneMoreItem    bool
	NoDuplicated       bool
}
