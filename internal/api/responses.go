package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
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
	Response
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Timestamp: time.Now().UTC(),
			Version:   "1.0.0",
		},
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Timestamp: time.Now().UTC(),
			Version:   "1.0.0",
		},
	})
}

func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
		Meta: &Meta{
			Timestamp: time.Now().UTC(),
			Version:   "1.0.0",
		},
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

func Paginated(c *gin.Context, data interface{}, page, perPage, totalCount int) {
	totalPages := (totalCount + perPage - 1) / perPage

	c.JSON(http.StatusOK, PaginatedResponse{
		Response: Response{
			Success: true,
			Data:    data,
			Meta: &Meta{
				Timestamp: time.Now().UTC(),
				Version:   "1.0.0",
			},
		},
		Pagination: &Pagination{
			Page:       page,
			PerPage:    perPage,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	})
}

type ValidationErrorResponse struct {
	Response
	ValidationErrors []APIValidationError `json:"validation_errors,omitempty"`
}

func ValidationError(c *gin.Context, errors []APIValidationError) {
	c.JSON(http.StatusBadRequest, ValidationErrorResponse{
		Response: Response{
			Success: false,
			Error:   "Validation failed",
			Meta: &Meta{
				Timestamp: time.Now().UTC(),
				Version:   "1.0.0",
			},
		},
		ValidationErrors: errors,
	})
}

type HealthResponse struct {
	Status     string            `json:"status"`
	Version    string            `json:"version"`
	Timestamp  time.Time         `json:"timestamp"`
	Uptime     string            `json:"uptime"`
	Components map[string]string `json:"components"`
}

func Health(c *gin.Context, health *HealthResponse) {
	c.JSON(http.StatusOK, health)
}

type MetricsResponse struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
}

func Metrics(c *gin.Context, metrics map[string]interface{}) {
	c.JSON(http.StatusOK, MetricsResponse{
		Timestamp: time.Now().UTC(),
		Metrics:   metrics,
	})
}

type StatusResponse struct {
	Status    string            `json:"status"`
	Services  map[string]string `json:"services"`
	Timestamp time.Time         `json:"timestamp"`
}

func Status(c *gin.Context, status *StatusResponse) {
	c.JSON(http.StatusOK, status)
}

type BulkOperationResponse struct {
	Response
	Processed int         `json:"processed"`
	Failed    int         `json:"failed"`
	Results   interface{} `json:"results,omitempty"`
}

func BulkOperation(c *gin.Context, processed, failed int, results interface{}) {
	c.JSON(http.StatusOK, BulkOperationResponse{
		Response: Response{
			Success: true,
			Meta: &Meta{
				Timestamp: time.Now().UTC(),
				Version:   "1.0.0",
			},
		},
		Processed: processed,
		Failed:    failed,
		Results:   results,
	})
}

type FileUploadResponse struct {
	Response
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	URL      string `json:"url,omitempty"`
}

func FileUpload(c *gin.Context, filename string, size int64, url string) {
	c.JSON(http.StatusOK, FileUploadResponse{
		Response: Response{
			Success: true,
			Meta: &Meta{
				Timestamp: time.Now().UTC(),
				Version:   "1.0.0",
			},
		},
		Filename: filename,
		Size:     size,
		URL:      url,
	})
}

type WebSocketResponse struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Channel string      `json:"channel,omitempty"`
}

func SendWebSocketMessage(c *gin.Context, msgType string, data interface{}) {
	response := WebSocketResponse{
		Type: msgType,
		Data: data,
	}

	c.JSON(http.StatusOK, response)
}

type RateLimitResponse struct {
	Response
	RetryAfter float64 `json:"retry_after"`
	Limit      int     `json:"limit"`
	Remaining  int     `json:"remaining"`
}

func RateLimitExceeded(c *gin.Context, retryAfter time.Duration, limit, remaining int) {
	c.JSON(http.StatusTooManyRequests, RateLimitResponse{
		Response: Response{
			Success: false,
			Error:   "Rate limit exceeded",
			Meta: &Meta{
				Timestamp: time.Now().UTC(),
				Version:   "1.0.0",
			},
		},
		RetryAfter: retryAfter.Seconds(),
		Limit:      limit,
		Remaining:  remaining,
	})
}
