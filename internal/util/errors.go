package util

import "fmt"

// Common error creation helpers for better DX (Developer Experience)

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string, identifier string) error {
	return &CustomError{
		Message:    fmt.Sprintf("%s not found", resource),
		ErrorCode:  USER_NOT_FOUND, // Can be made generic
		StatusCode: 404,
		Detail:     fmt.Sprintf("%s with identifier %s was not found", resource, identifier),
	}
}

// NewAlreadyExistsError creates an already exists error
func NewAlreadyExistsError(resource string, field string, value string) error {
	errorCode := USER_ALREADY_EXISTS
	if field == "email" {
		errorCode = EMAIL_ALREADY_EXISTS
	}

	return &CustomError{
		Message:    fmt.Sprintf("%s already exists", resource),
		ErrorCode:  errorCode,
		StatusCode: 400,
		Detail:     fmt.Sprintf("%s with %s '%s' already exists", resource, field, value),
	}
}

// NewValidationError creates a validation error
func NewValidationError(detail string) error {
	return &CustomError{
		Message:    "Validation failed",
		ErrorCode:  VALIDATION_ERROR,
		StatusCode: 400,
		Detail:     detail,
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, err error) error {
	return &CustomError{
		Message:    "Database operation failed",
		ErrorCode:  DATABASE_ERROR,
		StatusCode: 500,
		Detail:     fmt.Sprintf("failed to %s: %v", operation, err),
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(detail string) error {
	return &CustomError{
		Message:    "Unauthorized",
		ErrorCode:  UNAUTHORIZED,
		StatusCode: 401,
		Detail:     detail,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(detail string) error {
	return &CustomError{
		Message:    "Forbidden",
		ErrorCode:  FORBIDDEN,
		StatusCode: 403,
		Detail:     detail,
	}
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(field string, reason string) error {
	return &CustomError{
		Message:    "Invalid input",
		ErrorCode:  INVALID_INPUT,
		StatusCode: 400,
		Detail:     fmt.Sprintf("%s: %s", field, reason),
	}
}

// NewInternalError creates an internal server error
func NewInternalError(detail string) error {
	return &CustomError{
		Message:    "Internal server error",
		ErrorCode:  INTERNAL_SERVER_ERROR,
		StatusCode: 500,
		Detail:     detail,
	}
}

// IsCustomError checks if an error is a CustomError
func IsCustomError(err error) bool {
	_, ok := err.(*CustomError)
	return ok
}

// GetCustomError extracts CustomError from error
func GetCustomError(err error) (*CustomError, bool) {
	customErr, ok := err.(*CustomError)
	return customErr, ok
}
