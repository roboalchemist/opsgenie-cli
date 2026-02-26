package api

import "fmt"

// APIResponse is the generic OpsGenie API response wrapper.
type APIResponse struct {
	Result    string      `json:"result,omitempty"`
	Took      float64     `json:"took,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// PaginatedResponse wraps paginated list responses.
type PaginatedResponse struct {
	Data       interface{} `json:"data,omitempty"`
	Paging     *Paging     `json:"paging,omitempty"`
	Took       float64     `json:"took,omitempty"`
	RequestID  string      `json:"requestId,omitempty"`
}

// Paging contains pagination metadata.
type Paging struct {
	Next  string `json:"next,omitempty"`
	Prev  string `json:"prev,omitempty"`
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
}

// ErrorResponse represents an OpsGenie API error.
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("OpsGenie API error %d: %s", e.Code, e.Message)
}
