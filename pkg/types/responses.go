package types

import "time"

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	RequestID string    `json:"request_id,omitempty"`
}

type PaginatedResponse struct {
	APIResponse
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Components map[string]string `json:"components"`
}

type MetricsResponse struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
}

type ValidationResponse struct {
	Valid      bool          `json:"valid"`
	Confidence float64       `json:"confidence"`
	Reason     string        `json:"reason"`
	Expected   time.Duration `json:"expected_delay"`
	Actual     time.Duration `json:"actual_diff"`
}

type PropagationResponse struct {
	Source  string                       `json:"source"`
	Targets map[string]*PropagationResult `json:"targets"`
}

type NodeListResponse struct {
	Nodes []*Node `json:"nodes"`
	Total int     `json:"total"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type BulkOperationResponse struct {
	Processed int         `json:"processed"`
	Failed    int         `json:"failed"`
	Results   interface{} `json:"results,omitempty"`
}