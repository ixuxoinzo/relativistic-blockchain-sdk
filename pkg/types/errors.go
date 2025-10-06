package types

import "fmt"

type ErrorCode string

const (
	ErrInternal         ErrorCode = "INTERNAL_ERROR"
	ErrInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	ErrPermissionDenied ErrorCode = "PERMISSION_DENIED"
	ErrRateLimit        ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrNetwork          ErrorCode = "NETWORK_ERROR"
	ErrValidation       ErrorCode = "VALIDATION_ERROR"
	ErrTimeout          ErrorCode = "TIMEOUT"
	ErrUnavailable      ErrorCode = "SERVICE_UNAVAILABLE"
)

type RelativisticError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Err     error     `json:"-"`
}

func (e *RelativisticError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *RelativisticError) Unwrap() error {
	return e.Err
}

func NewError(code ErrorCode, message string) *RelativisticError {
	return &RelativisticError{
		Code:    code,
		Message: message,
	}
}

func NewErrorWithDetails(code ErrorCode, message, details string) *RelativisticError {
	return &RelativisticError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func WrapError(err error, code ErrorCode, message string) *RelativisticError {
	return &RelativisticError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func IsNotFound(err error) bool {
	if re, ok := err.(*RelativisticError); ok {
		return re.Code == ErrNotFound
	}
	return false
}

func IsPermissionDenied(err error) bool {
	if re, ok := err.(*RelativisticError); ok {
		return re.Code == ErrPermissionDenied
	}
	return false
}

func IsRateLimit(err error) bool {
	if re, ok := err.(*RelativisticError); ok {
		return re.Code == ErrRateLimit
	}
	return false
}