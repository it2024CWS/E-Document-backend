package util

// ErrorCode defines error code constants
type ErrorCode string

const (
	// Authentication & Authorization
	UNAUTHORIZED        ErrorCode = "UNAUTHORIZED"
	INCORRECT_PASSWORD  ErrorCode = "INCORRECT_PASSWORD"
	INVALID_CREDENTIALS ErrorCode = "INVALID_CREDENTIALS"
	TOKEN_EXPIRED       ErrorCode = "TOKEN_EXPIRED"
	INVALID_TOKEN       ErrorCode = "INVALID_TOKEN"
	FORBIDDEN           ErrorCode = "FORBIDDEN"

	// Validation errors
	VALIDATION_ERROR       ErrorCode = "VALIDATION_ERROR"
	MISSING_REQUIRED_FIELD ErrorCode = "MISSING_REQUIRED_FIELD"
	INVALID_INPUT          ErrorCode = "INVALID_INPUT"

	// Server errors
	INTERNAL_SERVER_ERROR ErrorCode = "INTERNAL_SERVER_ERROR"
	DATABASE_ERROR        ErrorCode = "DATABASE_ERROR"
	CONFIG_NOT_SET        ErrorCode = "CONFIG_NOT_SET"

	//USer errors
	USER_NOT_FOUND       ErrorCode = "USER_NOT_FOUND"
	USER_ALREADY_EXISTS  ErrorCode = "USER_ALREADY_EXISTS"
	EMAIL_ALREADY_EXISTS ErrorCode = "EMAIL_ALREADY_EXISTS"
)

// ErrorDetail represents detailed error information
type ErrorDetail struct {
	Detail string `json:"detail"`
}

// CustomError is a custom error type that includes error code and status code
type CustomError struct {
	Message    string
	ErrorCode  ErrorCode
	StatusCode int
	Detail     string
}

// Error implements the error interface
func (e *CustomError) Error() string {
	return e.Detail
}

// ErrorResponse creates a new CustomError
// Usage in service: return nil, util.ErrorResponse("Email already exists", util.EMAIL_ALREADY_EXISTS, 400, "email user@example.com is already in use")
// Usage in handler: return util.HandleError(c, util.ErrorResponse("Validation failed", util.INVALID_INPUT, 400, "Email is required"))
func ErrorResponse(message string, errorCode ErrorCode, statusCode int, detail string) error {
	return &CustomError{
		Message:    message,
		ErrorCode:  errorCode,
		StatusCode: statusCode,
		Detail:     detail,
	}
}
