package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents the standard API response envelope
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Data sends a successful JSON response with data payload
// Used for: List/Get operations that return data
func Data(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Message sends a successful JSON response with a message
// Used for: Operations that return confirmation messages (Start/Stop/Delete)
func Message(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
	})
}

// Created sends a JSON response with status 201 Created and data payload
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// Error sends a JSON error response with a custom status code
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Success: false,
		Message: message,
	})
}

// BadRequest sends a JSON error response with status 400 Bad Request
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// NotFound sends a JSON error response with status 404 Not Found
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalServerError sends a JSON error response with status 500 Internal Server Error
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// Legacy compatibility functions (deprecated, will be removed in future versions)
// These are kept temporarily to avoid breaking changes during gradual migration

// Success sends a JSON response with status 200 OK
// Deprecated: Use Data() or Message() instead for clearer semantics
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// WithMessage adds a message field to the data map if it's a gin.H or map[string]interface{}
// Deprecated: Use Message() instead
func WithMessage(data gin.H, message string) gin.H {
	data["message"] = message
	return data
}
