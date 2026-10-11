package controlplane

import (
	"errors"
	"fmt"
)

// Service layer errors
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrProjectNotFound     = errors.New("project not found")
	ErrApplicationNotFound = errors.New("application not found")
	ErrDeploymentNotFound  = errors.New("deployment not found")
	ErrInvalidRelationship = errors.New("invalid resource relationship")
	ErrInvalidStatus       = errors.New("invalid status transition")
	ErrResourceConflict    = errors.New("resource already exists")
	ErrValidation          = errors.New("validation failed")
	ErrInternal            = errors.New("internal server error")
)

// ValidationError represents a specific validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error codes for HTTP responses
const (
	CodeValidationFailed    = "VALIDATION_FAILED"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeInvalidRelationship = "INVALID_RELATIONSHIP"
	CodeInternal            = "INTERNAL"
)

// ToHTTPError converts service errors to HTTP error responses
func ToHTTPError(err error) HTTPError {
	if err == nil {
		return HTTPError{}
	}

	switch {
	case errors.Is(err, ErrUserNotFound),
		errors.Is(err, ErrProjectNotFound),
		errors.Is(err, ErrApplicationNotFound),
		errors.Is(err, ErrDeploymentNotFound):
		return HTTPError{
			Code:    CodeNotFound,
			Message: "Resource was not found",
		}
	case errors.Is(err, ErrInvalidRelationship):
		return HTTPError{
			Code:    CodeInvalidRelationship,
			Message: "Invalid resource relationship",
		}
	case errors.Is(err, ErrInvalidStatus):
		return HTTPError{
			Code:    CodeValidationFailed,
			Message: "Invalid status transition",
		}
	case errors.Is(err, ErrResourceConflict):
		return HTTPError{
			Code:    CodeConflict,
			Message: "Resource already exists",
		}
	case errors.Is(err, ErrValidation):
		return HTTPError{
			Code:    CodeValidationFailed,
			Message: "Validation failed",
		}
	default:
		return HTTPError{
			Code:    CodeInternal,
			Message: "Internal server error",
		}
	}
}
