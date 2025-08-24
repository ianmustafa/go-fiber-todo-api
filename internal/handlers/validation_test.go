package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"go-fiber/internal/config"
	"go-fiber/internal/mocks"
	"go-fiber/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupValidationTest creates a fresh handler and app for each test to avoid mock contamination
func setupValidationTest() (*fiber.App, *mocks.MockTodoRepository) {
	mockRepo := new(mocks.MockTodoRepository)
	logger := config.NewTestLogger()
	validator := validator.New()
	handler := NewTodoHandler(mockRepo, validator, logger)

	app := fiber.New()
	authMiddleware := func(c *fiber.Ctx) error {
		c.Locals("userID", "test-user-id")
		c.Locals("username", "testuser")
		return c.Next()
	}

	api := app.Group("/api/v1")
	handler.RegisterRoutes(api, authMiddleware)

	return app, mockRepo
}

func TestQueryParameterValidation(t *testing.T) {
	t.Run("valid query parameters", func(t *testing.T) {
		app, mockRepo := setupValidationTest()

		// Mock successful response
		mockRepo.On("GetByUserID", mock.Anything, "test-user-id", 5, 10).Return([]*models.Todo{}, int64(0), nil)

		req := httptest.NewRequest("GET", "/api/v1/todos?limit=5&offset=10", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid limit - too high", func(t *testing.T) {
		app, _ := setupValidationTest()

		req := httptest.NewRequest("GET", "/api/v1/todos?limit=200", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "Validation Error", response["error"])
	})

	t.Run("invalid status", func(t *testing.T) {
		app, _ := setupValidationTest()

		req := httptest.NewRequest("GET", "/api/v1/todos?status=invalid_status", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "Validation Error", response["error"])
	})

	t.Run("invalid priority", func(t *testing.T) {
		app, _ := setupValidationTest()

		req := httptest.NewRequest("GET", "/api/v1/todos?priority=invalid_priority", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "Validation Error", response["error"])
	})
}
