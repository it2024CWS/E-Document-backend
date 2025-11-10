package util

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response represents a standard API response structure
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	ErrorCode ErrorCode   `json:"error_code"`
	Data      interface{} `json:"data,omitempty"`
}

// SuccessResponse returns a successful response
func SuccessResponse(c echo.Context, statusCode int, message string, data interface{}) error {
	return c.JSON(statusCode, Response{
		Success:   true,
		Message:   message,
		ErrorCode: "",
		Data:      data,
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

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	CurrentPage  int `json:"currentPage"`
	TotalPages   int `json:"totalPages"`
	TotalItems   int `json:"totalItems"`
	ItemsPerPage int `json:"itemsPerPage"`
}

// PaginatedData wraps items with pagination info
type PaginatedData struct {
	Items      interface{}    `json:"items"`
	Pagination PaginationInfo `json:"pagination"`
}

// OKResponseWithPagination returns a 200 OK response with pagination
func OKResponseWithPagination(c echo.Context, message string, items interface{}, pagination PaginationInfo) error {
	data := PaginatedData{
		Items:      items,
		Pagination: pagination,
	}
	return SuccessResponse(c, http.StatusOK, message, data)
}

// HandleError handles error and returns appropriate response
// If error is CustomError, use its info; otherwise return 500
func HandleError(c echo.Context, err error) error {
	if customErr, ok := err.(*CustomError); ok {
		// Use CustomError info
		data := ErrorDetail{Detail: customErr.Detail}
		return c.JSON(customErr.StatusCode, Response{
			Success:   false,
			Message:   customErr.Message,
			ErrorCode: customErr.ErrorCode,
			Data:      data,
		})
	}

	// Regular error - return 500
	data := ErrorDetail{Detail: err.Error()}
	return c.JSON(http.StatusInternalServerError, Response{
		Success:   false,
		Message:   "Internal server error",
		ErrorCode: INTERNAL_SERVER_ERROR,
		Data:      data,
	})
}
