package errors

import "fmt"

// AppError represents an application error with context
type AppError struct {
	Code    string // Error code for programmatic handling
	Message string // Human-readable error message
	Status  int    // HTTP status code
	Err     error  // Underlying error (optional)
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap allows errors.Is and errors.As to work
func (e *AppError) Unwrap() error {
	return e.Err
}

// Common error codes
const (
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeProviderFailure = "PROVIDER_FAILURE"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeInternal        = "INTERNAL_ERROR"
	ErrCodeTimeout         = "TIMEOUT"
	ErrCodeRateLimit       = "RATE_LIMIT_EXCEEDED"
)

// Predefined errors
var (
	ErrInvalidRequest     = &AppError{Code: ErrCodeValidation, Message: "Invalid request", Status: 400}
	ErrProviderUnavailable = &AppError{Code: ErrCodeProviderFailure, Message: "Provider unavailable", Status: 503}
	ErrNoFlightsFound     = &AppError{Code: ErrCodeNotFound, Message: "No flights found", Status: 404}
	ErrInternalServer     = &AppError{Code: ErrCodeInternal, Message: "Internal server error", Status: 500}
	ErrRequestTimeout     = &AppError{Code: ErrCodeTimeout, Message: "Request timeout", Status: 504}
	ErrRateLimitExceeded  = &AppError{Code: ErrCodeRateLimit, Message: "Rate limit exceeded", Status: 429}
)

// NewAppError creates a new AppError with custom message
func NewAppError(code, message string, status int, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Err:     err,
	}
}
