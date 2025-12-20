package util

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response represents a standard API response structure
type Response struct {
	Success    bool           `json:"success"`
	Message    string         `json:"message,omitempty"`
	ErrorCode  ErrorCode      `json:"error_code"`
	Data       interface{}    `json:"data,omitempty"`
	Pagination PaginationInfo `json:"pagination,omitempty"`
}

// SuccessResponse returns a successful response

func SuccessResponse(c echo.Context, statusCode int, message string, data interface{}, pagination PaginationInfo) error {
	return c.JSON(statusCode, Response{
		Success:    true,
		Message:    message,
		ErrorCode:  "",
		Data:       data,
		Pagination: pagination,
	})
}

// OKResponse returns a success response with optional status code (default 200)
func OKResponse(c echo.Context, message string, data interface{}, statusCode ...int) error {
	code := http.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	return SuccessResponse(c, code, message, data, PaginationInfo{})
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
	Items interface{} `json:"items"`
}

// OKResponseWithPagination returns a 200 OK response with pagination
func OKResponseWithPagination(c echo.Context, message string, items interface{}, pagination PaginationInfo) error {
	data := PaginatedData{
		Items: items,
	}
	return SuccessResponse(c, http.StatusOK, message, data, pagination)
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
