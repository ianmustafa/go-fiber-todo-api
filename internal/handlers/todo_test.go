package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"go-fiber/internal/config"
	"go-fiber/internal/mocks"
	"go-fiber/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTodoHandler() (*TodoHandler, *mocks.MockTodoRepository) {
	mockRepo := new(mocks.MockTodoRepository)
	logger := config.NewTestLogger()
	validator := validator.New()
	handler := NewTodoHandler(mockRepo, validator, logger)
	return handler, mockRepo
}

func setupFiberApp(handler *TodoHandler) *fiber.App {
	app := fiber.New()

	// Add middleware to set user context for testing
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", "test-user-id")
		c.Locals("username", "testuser")
		return c.Next()
	})

	// Register routes
	api := app.Group("/api/v1")
	todos := api.Group("/todos")

	todos.Post("/", handler.CreateTodo)
	todos.Get("/", handler.GetTodos)
	todos.Get("/:id", handler.GetTodo)
	todos.Put("/:id", handler.UpdateTodo)
	todos.Delete("/:id", handler.DeleteTodo)

	return app
}

func TestTodoHandler_CreateTodo(t *testing.T) {
	handler, mockRepo := setupTodoHandler()
	app := setupFiberApp(handler)

	t.Run("successful todo creation", func(t *testing.T) {
		// Arrange
		reqBody := models.CreateTodoRequest{
			Title:       "Test Todo",
			Description: "Test Description",
		}

		expectedTodo := &models.Todo{
			ID:          "todo-id",
			UserID:      "test-user-id",
			Title:       "Test Todo",
			Description: "Test Description",
			Status:      models.TodoStatusPending,
			Priority:    models.TodoPriorityMedium,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Todo")).Return(expectedTodo, nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/todos/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)

		var response models.Todo
		json.NewDecoder(resp.Body).Decode(&response)

		assert.Equal(t, "Test Todo", response.Title)
		assert.Equal(t, "Test Description", response.Description)
		assert.Equal(t, models.TodoStatusPending, response.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("POST", "/api/v1/todos/", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("validation error - empty title", func(t *testing.T) {
		// Arrange
		reqBody := models.CreateTodoRequest{
			Title:       "", // Empty title should fail validation
			Description: "Test Description",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/todos/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestTodoHandler_GetTodos(t *testing.T) {
	handler, mockRepo := setupTodoHandler()
	app := setupFiberApp(handler)

	t.Run("successful get todos", func(t *testing.T) {
		// Arrange
		expectedTodos := []*models.Todo{
			{
				ID:          "todo-1",
				UserID:      "test-user-id",
				Title:       "Todo 1",
				Description: "Description 1",
				Status:      models.TodoStatusPending,
				Priority:    models.TodoPriorityMedium,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "todo-2",
				UserID:      "test-user-id",
				Title:       "Todo 2",
				Description: "Description 2",
				Status:      models.TodoStatusCompleted,
				Priority:    models.TodoPriorityHigh,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mockRepo.On("GetByUserID", mock.Anything, "test-user-id", 10, 0).Return(expectedTodos, int64(2), nil)

		req := httptest.NewRequest("GET", "/api/v1/todos/", nil)

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var response models.TodoListResponse
		json.NewDecoder(resp.Body).Decode(&response)

		assert.Len(t, response.Todos, 2)
		assert.Equal(t, int64(2), response.Total)
		assert.Equal(t, 10, response.Limit)
		assert.Equal(t, 0, response.Offset)

		mockRepo.AssertExpectations(t)
	})

	t.Run("get todos with pagination", func(t *testing.T) {
		// Arrange
		expectedTodos := []*models.Todo{
			{
				ID:          "todo-3",
				UserID:      "test-user-id",
				Title:       "Todo 3",
				Description: "Description 3",
				Status:      models.TodoStatusPending,
				Priority:    models.TodoPriorityLow,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mockRepo.On("GetByUserID", mock.Anything, "test-user-id", 5, 5).Return(expectedTodos, int64(6), nil)

		req := httptest.NewRequest("GET", "/api/v1/todos/?limit=5&offset=5", nil)

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var response models.TodoListResponse
		json.NewDecoder(resp.Body).Decode(&response)

		assert.Len(t, response.Todos, 1)
		assert.Equal(t, int64(6), response.Total)
		assert.Equal(t, 5, response.Limit)
		assert.Equal(t, 5, response.Offset)

		mockRepo.AssertExpectations(t)
	})
}

func TestTodoHandler_GetTodo(t *testing.T) {
	handler, mockRepo := setupTodoHandler()
	app := setupFiberApp(handler)

	t.Run("successful get todo", func(t *testing.T) {
		// Arrange
		expectedTodo := &models.Todo{
			ID:          "todo-1",
			UserID:      "test-user-id",
			Title:       "Test Todo",
			Description: "Test Description",
			Status:      models.TodoStatusPending,
			Priority:    models.TodoPriorityMedium,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("GetByID", mock.Anything, "todo-1").Return(expectedTodo, nil)

		req := httptest.NewRequest("GET", "/api/v1/todos/todo-1", nil)

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var response models.Todo
		json.NewDecoder(resp.Body).Decode(&response)

		assert.Equal(t, "Test Todo", response.Title)
		assert.Equal(t, "Test Description", response.Description)

		mockRepo.AssertExpectations(t)
	})

	t.Run("todo not found", func(t *testing.T) {
		// Arrange
		mockRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

		req := httptest.NewRequest("GET", "/api/v1/todos/nonexistent", nil)

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode) // Handler returns 500 for generic errors

		mockRepo.AssertExpectations(t)
	})
}

func TestTodoHandler_UpdateTodo(t *testing.T) {
	handler, mockRepo := setupTodoHandler()
	app := setupFiberApp(handler)

	t.Run("successful todo update", func(t *testing.T) {
		// Arrange
		reqBody := models.UpdateTodoRequest{
			Title:       "Updated Todo",
			Description: "Updated Description",
			Status:      models.TodoStatusCompleted,
		}

		existingTodo := &models.Todo{
			ID:          "todo-1",
			UserID:      "test-user-id",
			Title:       "Original Todo",
			Description: "Original Description",
			Status:      models.TodoStatusPending,
			Priority:    models.TodoPriorityMedium,
			CreatedAt:   time.Now().Add(-time.Hour),
			UpdatedAt:   time.Now().Add(-time.Hour),
		}

		updatedTodo := &models.Todo{
			ID:          "todo-1",
			UserID:      "test-user-id",
			Title:       "Updated Todo",
			Description: "Updated Description",
			Status:      models.TodoStatusCompleted,
			Priority:    models.TodoPriorityMedium,
			CreatedAt:   existingTodo.CreatedAt,
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("GetByID", mock.Anything, "todo-1").Return(existingTodo, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Todo")).Return(updatedTodo, nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/api/v1/todos/todo-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var response models.Todo
		json.NewDecoder(resp.Body).Decode(&response)

		assert.Equal(t, "Updated Todo", response.Title)
		assert.Equal(t, "Updated Description", response.Description)
		assert.Equal(t, models.TodoStatusCompleted, response.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("todo not found", func(t *testing.T) {
		// Arrange
		reqBody := models.UpdateTodoRequest{
			Title: "Updated Todo",
		}

		mockRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/api/v1/todos/nonexistent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode) // Handler returns 500 for generic errors

		mockRepo.AssertExpectations(t)
	})
}

func TestTodoHandler_DeleteTodo(t *testing.T) {
	handler, mockRepo := setupTodoHandler()
	app := setupFiberApp(handler)

	t.Run("successful todo deletion", func(t *testing.T) {
		// Arrange
		existingTodo := &models.Todo{
			ID:          "todo-1",
			UserID:      "test-user-id",
			Title:       "Todo to Delete",
			Description: "Description",
			Status:      models.TodoStatusPending,
			Priority:    models.TodoPriorityMedium,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("GetByID", mock.Anything, "todo-1").Return(existingTodo, nil)
		mockRepo.On("Delete", mock.Anything, "todo-1").Return(nil)

		req := httptest.NewRequest("DELETE", "/api/v1/todos/todo-1", nil)

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("todo not found", func(t *testing.T) {
		// Arrange
		mockRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

		req := httptest.NewRequest("DELETE", "/api/v1/todos/nonexistent", nil)

		// Act
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode) // Handler returns 500 for generic errors

		mockRepo.AssertExpectations(t)
	})
}
