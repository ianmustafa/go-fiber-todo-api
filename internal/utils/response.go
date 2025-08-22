package utils

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	TotalPages int         `json:"total_pages"`
	Page       int         `json:"page"`
}

// SendError sends an error response
func SendError(c *fiber.Ctx, statusCode int, message string, details ...interface{}) error {
	response := ErrorResponse{
		Error:   fiber.ErrBadRequest.Message,
		Message: message,
	}

	if len(details) > 0 {
		response.Details = details[0]
	}

	// Set appropriate error message based on status code
	switch statusCode {
	case fiber.StatusBadRequest:
		response.Error = "Bad Request"
	case fiber.StatusUnauthorized:
		response.Error = "Unauthorized"
	case fiber.StatusForbidden:
		response.Error = "Forbidden"
	case fiber.StatusNotFound:
		response.Error = "Not Found"
	case fiber.StatusConflict:
		response.Error = "Conflict"
	case fiber.StatusUnprocessableEntity:
		response.Error = "Unprocessable Entity"
	case fiber.StatusInternalServerError:
		response.Error = "Internal Server Error"
	default:
		response.Error = "Error"
	}

	return c.Status(statusCode).JSON(response)
}

// SendSuccess sends a success response
func SendSuccess(c *fiber.Ctx, message string, data ...interface{}) error {
	response := SuccessResponse{
		Message: message,
	}

	if len(data) > 0 {
		response.Data = data[0]
	}

	return c.JSON(response)
}

// SendData sends data response without message
func SendData(c *fiber.Ctx, data interface{}) error {
	return c.JSON(data)
}

// SendPaginated sends a paginated response
func SendPaginated(c *fiber.Ctx, data interface{}, total int64, limit, offset int) error {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	page := (offset / limit) + 1

	response := PaginatedResponse{
		Data:       data,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPages,
		Page:       page,
	}

	return c.JSON(response)
}

// SendCreated sends a created response
func SendCreated(c *fiber.Ctx, message string, data interface{}) error {
	response := SuccessResponse{
		Message: message,
		Data:    data,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// SendNoContent sends a no content response
func SendNoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// SendValidationError sends a validation error response
func SendValidationError(c *fiber.Ctx, errors []string) error {
	return SendError(c, fiber.StatusBadRequest, "Validation failed", errors)
}

// SendUnauthorized sends an unauthorized response
func SendUnauthorized(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Authentication required"
	}
	return SendError(c, fiber.StatusUnauthorized, message)
}

// SendForbidden sends a forbidden response
func SendForbidden(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Access denied"
	}
	return SendError(c, fiber.StatusForbidden, message)
}

// SendNotFound sends a not found response
func SendNotFound(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Resource not found"
	}
	return SendError(c, fiber.StatusNotFound, message)
}

// SendConflict sends a conflict response
func SendConflict(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Resource already exists"
	}
	return SendError(c, fiber.StatusConflict, message)
}

// SendInternalError sends an internal server error response
func SendInternalError(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Internal server error"
	}
	return SendError(c, fiber.StatusInternalServerError, message)
}
