package util

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse returns a successful response
func SuccessResponse(c echo.Context, statusCode int, message string, data interface{}) error {
	return c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse returns an error response
func ErrorResponse(c echo.Context, statusCode int, message string, err error) error {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	return c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// CreatedResponse returns a 201 Created response
func CreatedResponse(c echo.Context, message string, data interface{}) error {
	return SuccessResponse(c, http.StatusCreated, message, data)
}

// OKResponse returns a 200 OK response
func OKResponse(c echo.Context, message string, data interface{}) error {
	return SuccessResponse(c, http.StatusOK, message, data)
}

// BadRequestResponse returns a 400 Bad Request response
func BadRequestResponse(c echo.Context, message string, err error) error {
	return ErrorResponse(c, http.StatusBadRequest, message, err)
}

// NotFoundResponse returns a 404 Not Found response
func NotFoundResponse(c echo.Context, message string, err error) error {
	return ErrorResponse(c, http.StatusNotFound, message, err)
}

// InternalServerErrorResponse returns a 500 Internal Server Error response
func InternalServerErrorResponse(c echo.Context, message string, err error) error {
	return ErrorResponse(c, http.StatusInternalServerError, message, err)
}

// UnauthorizedResponse returns a 401 Unauthorized response
func UnauthorizedResponse(c echo.Context, message string, err error) error {
	return ErrorResponse(c, http.StatusUnauthorized, message, err)
}
