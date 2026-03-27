package httputil

// APIResponse is the standard API response.
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PagedResponse is the paginated API response.
type PagedResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination holds pagination info.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// OK returns a success response.
func OK(data interface{}) APIResponse {
	return APIResponse{Code: 0, Message: "ok", Data: data}
}

// Error returns an error response.
func Error(code int, message string) APIResponse {
	return APIResponse{Code: code, Message: message}
}
