package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	ErrValidation = "VALIDATION_ERROR"
	ErrUnauthorized = "UNAUTHORIZED"
	ErrForbidden    = "FORBIDDEN"
	ErrNotFound     = "NOT_FOUND"
	ErrConflict     = "CONFLICT"
	ErrInternal     = "INTERNAL_ERROR"
)

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func List(c *gin.Context, data any, pagination Pagination) {
	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"pagination": pagination,
	})
}

func Error(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{
		"error": errorBody{
			Code:    code,
			Message: message,
		},
	})
}

func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, ErrValidation, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, ErrUnauthorized, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, ErrNotFound, message)
}

func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, ErrConflict, message)
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, ErrInternal, "internal server error")
}
