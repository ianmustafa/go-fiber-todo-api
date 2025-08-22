package handlers

import (
	"strconv"

	"go-fiber/internal/middleware"
	"go-fiber/internal/models"
	"go-fiber/internal/repository/interfaces"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// TodoHandler handles todo-related HTTP requests
type TodoHandler struct {
	todoRepo  interfaces.TodoRepository
	validator *validator.Validate
	logger    zerolog.Logger
}

// NewTodoHandler creates a new todo handler
func NewTodoHandler(todoRepo interfaces.TodoRepository, validator *validator.Validate, logger zerolog.Logger) *TodoHandler {
	return &TodoHandler{
		todoRepo:  todoRepo,
		validator: validator,
		logger:    logger,
	}
}

// CreateTodo handles todo creation
// @Summary Create a new todo
// @Description Create a new todo item for the authenticated user
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateTodoRequest true "Create todo request"
// @Success 201 {object} models.Todo
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /todos [post]
func (h *TodoHandler) CreateTodo(c *fiber.Ctx) error {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}

	var req models.CreateTodoRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse create todo request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error().Err(err).Msg("Create todo request validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation Error",
			"message": "Invalid input data",
			"details": err.Error(),
		})
	}

	// Create todo
	todo := &models.Todo{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	}

	createdTodo, err := h.todoRepo.Create(c.Context(), todo)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create todo")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create todo",
		})
	}

	h.logger.Info().Str("todo_id", createdTodo.ID).Str("user_id", userID).Msg("Todo created successfully")
	return c.Status(fiber.StatusCreated).JSON(createdTodo)
}

// GetTodos handles getting user's todos with pagination
// @Summary Get user's todos
// @Description Get todos for the authenticated user with pagination and filtering
// @Tags todos
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of todos to return" default(10)
// @Param offset query int false "Number of todos to skip" default(0)
// @Param status query string false "Filter by status" Enums(pending, in_progress, completed)
// @Param priority query string false "Filter by priority" Enums(low, medium, high)
// @Success 200 {object} models.TodoListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /todos [get]
func (h *TodoHandler) GetTodos(c *fiber.Ctx) error {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	status := c.Query("status")
	priority := c.Query("priority")

	// Validate limit and offset
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var todos []*models.Todo
	var total int64
	var err error

	// Filter by status or priority if provided
	if status != "" && models.IsValidStatus(status) {
		todos, total, err = h.todoRepo.GetByStatus(c.Context(), userID, status, limit, offset)
	} else if priority != "" && models.IsValidPriority(priority) {
		todos, total, err = h.todoRepo.GetByPriority(c.Context(), userID, priority, limit, offset)
	} else {
		todos, total, err = h.todoRepo.GetByUserID(c.Context(), userID, limit, offset)
	}

	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get todos")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get todos",
		})
	}

	response := &models.TodoListResponse{
		Todos:  todos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	return c.JSON(response)
}

// GetTodo handles getting a specific todo
// @Summary Get a todo by ID
// @Description Get a specific todo by its ID
// @Tags todos
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo ID"
// @Success 200 {object} models.Todo
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /todos/{id} [get]
func (h *TodoHandler) GetTodo(c *fiber.Ctx) error {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}

	// Get todo ID from params
	todoID := c.Params("id")
	if todoID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Todo ID is required",
		})
	}

	// Get todo
	todo, err := h.todoRepo.GetByID(c.Context(), todoID)
	if err != nil {
		if err.Error() == "todo not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Todo not found",
			})
		}
		h.logger.Error().Err(err).Str("todo_id", todoID).Msg("Failed to get todo")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get todo",
		})
	}

	// Check if todo belongs to the authenticated user
	if todo.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Todo not found",
		})
	}

	return c.JSON(todo)
}

// UpdateTodo handles todo updates
// @Summary Update a todo
// @Description Update a specific todo by its ID
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo ID"
// @Param request body models.UpdateTodoRequest true "Update todo request"
// @Success 200 {object} models.Todo
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /todos/{id} [put]
func (h *TodoHandler) UpdateTodo(c *fiber.Ctx) error {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}

	// Get todo ID from params
	todoID := c.Params("id")
	if todoID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Todo ID is required",
		})
	}

	var req models.UpdateTodoRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse update todo request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error().Err(err).Msg("Update todo request validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation Error",
			"message": "Invalid input data",
			"details": err.Error(),
		})
	}

	// Get existing todo to verify ownership
	existingTodo, err := h.todoRepo.GetByID(c.Context(), todoID)
	if err != nil {
		if err.Error() == "todo not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Todo not found",
			})
		}
		h.logger.Error().Err(err).Str("todo_id", todoID).Msg("Failed to get todo for update")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get todo",
		})
	}

	// Check if todo belongs to the authenticated user
	if existingTodo.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Todo not found",
		})
	}

	// Update todo fields
	if req.Title != "" {
		existingTodo.Title = req.Title
	}
	if req.Description != "" {
		existingTodo.Description = req.Description
	}
	if req.Status != "" {
		existingTodo.Status = req.Status
	}
	if req.Priority != "" {
		existingTodo.Priority = req.Priority
	}
	if req.DueDate != nil {
		existingTodo.DueDate = req.DueDate
	}

	// Update todo
	updatedTodo, err := h.todoRepo.Update(c.Context(), existingTodo)
	if err != nil {
		h.logger.Error().Err(err).Str("todo_id", todoID).Msg("Failed to update todo")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update todo",
		})
	}

	h.logger.Info().Str("todo_id", todoID).Str("user_id", userID).Msg("Todo updated successfully")
	return c.JSON(updatedTodo)
}

// DeleteTodo handles todo deletion
// @Summary Delete a todo
// @Description Delete a specific todo by its ID
// @Tags todos
// @Security BearerAuth
// @Param id path string true "Todo ID"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /todos/{id} [delete]
func (h *TodoHandler) DeleteTodo(c *fiber.Ctx) error {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}

	// Get todo ID from params
	todoID := c.Params("id")
	if todoID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Todo ID is required",
		})
	}

	// Get existing todo to verify ownership
	existingTodo, err := h.todoRepo.GetByID(c.Context(), todoID)
	if err != nil {
		if err.Error() == "todo not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Todo not found",
			})
		}
		h.logger.Error().Err(err).Str("todo_id", todoID).Msg("Failed to get todo for deletion")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get todo",
		})
	}

	// Check if todo belongs to the authenticated user
	if existingTodo.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Todo not found",
		})
	}

	// Delete todo
	if err := h.todoRepo.Delete(c.Context(), todoID); err != nil {
		h.logger.Error().Err(err).Str("todo_id", todoID).Msg("Failed to delete todo")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete todo",
		})
	}

	h.logger.Info().Str("todo_id", todoID).Str("user_id", userID).Msg("Todo deleted successfully")
	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateTodoStatus handles todo status updates
// @Summary Update todo status
// @Description Update the status of a specific todo
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo ID"
// @Param request body models.UpdateTodoStatusRequest true "Update status request"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /todos/{id}/status [patch]
func (h *TodoHandler) UpdateTodoStatus(c *fiber.Ctx) error {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}

	// Get todo ID from params
	todoID := c.Params("id")
	if todoID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Todo ID is required",
		})
	}

	var req models.UpdateTodoStatusRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse update status request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error().Err(err).Msg("Update status request validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation Error",
			"message": "Invalid input data",
			"details": err.Error(),
		})
	}

	// Get existing todo to verify ownership
	existingTodo, err := h.todoRepo.GetByID(c.Context(), todoID)
	if err != nil {
		if err.Error() == "todo not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Todo not found",
			})
		}
		h.logger.Error().Err(err).Str("todo_id", todoID).Msg("Failed to get todo for status update")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get todo",
		})
	}

	// Check if todo belongs to the authenticated user
	if existingTodo.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Todo not found",
		})
	}

	// Update status
	if err := h.todoRepo.UpdateStatus(c.Context(), todoID, req.Status); err != nil {
		h.logger.Error().Err(err).Str("todo_id", todoID).Msg("Failed to update todo status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update todo status",
		})
	}

	h.logger.Info().Str("todo_id", todoID).Str("status", req.Status).Str("user_id", userID).Msg("Todo status updated successfully")
	return c.JSON(fiber.Map{
		"message": "Todo status updated successfully",
		"status":  req.Status,
	})
}

// GetOverdueTodos handles getting overdue todos
// @Summary Get overdue todos
// @Description Get overdue todos for the authenticated user
// @Tags todos
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of todos to return" default(10)
// @Param offset query int false "Number of todos to skip" default(0)
// @Success 200 {object} models.TodoListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /todos/overdue [get]
func (h *TodoHandler) GetOverdueTodos(c *fiber.Ctx) error {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// Validate limit and offset
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get overdue todos
	todos, total, err := h.todoRepo.GetOverdue(c.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get overdue todos")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get overdue todos",
		})
	}

	response := &models.TodoListResponse{
		Todos:  todos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	return c.JSON(response)
}

// SearchTodos handles todo search
// @Summary Search todos
// @Description Search todos by title and description
// @Tags todos
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param limit query int false "Number of todos to return" default(10)
// @Param offset query int false "Number of todos to skip" default(0)
// @Success 200 {object} models.TodoListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /todos/search [get]
func (h *TodoHandler) SearchTodos(c *fiber.Ctx) error {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}

	// Get search query
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Search query is required",
		})
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// Validate limit and offset
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Search todos
	todos, total, err := h.todoRepo.Search(c.Context(), userID, query, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Str("query", query).Msg("Failed to search todos")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to search todos",
		})
	}

	response := &models.TodoListResponse{
		Todos:  todos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	return c.JSON(response)
}

// GetTodoStats handles getting todo statistics
// @Summary Get todo statistics
// @Description Get todo statistics for the authenticated user
// @Tags todos
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /todos/stats [get]
func (h *TodoHandler) GetTodoStats(c *fiber.Ctx) error {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}

	// Get todo statistics
	stats, err := h.todoRepo.CountByStatus(c.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get todo statistics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get todo statistics",
		})
	}

	return c.JSON(fiber.Map{
		"stats": stats,
	})
}

// RegisterRoutes registers todo routes
func (h *TodoHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	todos := router.Group("/todos", authMiddleware)

	// CRUD operations
	todos.Post("/", h.CreateTodo)
	todos.Get("/", h.GetTodos)
	todos.Get("/:id", h.GetTodo)
	todos.Put("/:id", h.UpdateTodo)
	todos.Delete("/:id", h.DeleteTodo)

	// Status operations
	todos.Patch("/:id/status", h.UpdateTodoStatus)

	// Special operations
	todos.Get("/overdue", h.GetOverdueTodos)
	todos.Get("/search", h.SearchTodos)
	todos.Get("/stats", h.GetTodoStats)
}
