package requests

type TestNotificationRequest struct{}

type PaginationRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	SortBy   string `json:"sortBy"`
	Order    string `json:"order" validate:"oneof=ASC DESC asc desc ''"`
}
