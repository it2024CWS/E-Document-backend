package util

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct validates a struct and returns formatted error messages
func ValidateStruct(data interface{}) error {
	err := validate.Struct(data)
	if err == nil {
		return nil
	}

	// Format validation errors
	var errorMessages []string
	for _, err := range err.(validator.ValidationErrors) {
		errorMessages = append(errorMessages, formatValidationError(err))
	}

	return ErrorResponse(
		"Validation failed",
		INVALID_INPUT,
		400,
		strings.Join(errorMessages, "; "),
	)
}

// formatValidationError formats a single validation error into a user-friendly message
func formatValidationError(err validator.FieldError) string {
	field := err.Field()

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, err.Param())
	case "e164":
		return fmt.Sprintf("%s must be a valid phone number in E.164 format (e.g., +66812345678)", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, err.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, err.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// GetValidator returns the validator instance for custom validations
func GetValidator() *validator.Validate {
	return validate
}
