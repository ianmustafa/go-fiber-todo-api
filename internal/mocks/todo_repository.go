package mocks

import (
	"context"

	"go-fiber/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockTodoRepository is a mock implementation of the TodoRepository interface
type MockTodoRepository struct {
	mock.Mock
}

// Create creates a new todo
func (m *MockTodoRepository) Create(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
	args := m.Called(ctx, todo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Todo), args.Error(1)
}

// GetByID retrieves a todo by ID
func (m *MockTodoRepository) GetByID(ctx context.Context, id string) (*models.Todo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Todo), args.Error(1)
}

// GetByUserID retrieves all todos for a specific user
func (m *MockTodoRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Todo, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Todo), args.Get(1).(int64), args.Error(2)
}

// Update updates an existing todo
func (m *MockTodoRepository) Update(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
	args := m.Called(ctx, todo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Todo), args.Error(1)
}

// Delete soft deletes a todo
func (m *MockTodoRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// UpdateStatus updates the status of a todo
func (m *MockTodoRepository) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// GetByStatus retrieves todos by user ID and status
func (m *MockTodoRepository) GetByStatus(ctx context.Context, userID, status string, limit, offset int) ([]*models.Todo, int64, error) {
	args := m.Called(ctx, userID, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Todo), args.Get(1).(int64), args.Error(2)
}

// GetByPriority retrieves todos by user ID and priority
func (m *MockTodoRepository) GetByPriority(ctx context.Context, userID, priority string, limit, offset int) ([]*models.Todo, int64, error) {
	args := m.Called(ctx, userID, priority, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Todo), args.Get(1).(int64), args.Error(2)
}

// GetOverdue retrieves overdue todos
func (m *MockTodoRepository) GetOverdue(ctx context.Context, userID string, limit, offset int) ([]*models.Todo, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Todo), args.Get(1).(int64), args.Error(2)
}

// GetUpcoming retrieves upcoming todos
func (m *MockTodoRepository) GetUpcoming(ctx context.Context, userID string, days int, limit, offset int) ([]*models.Todo, int64, error) {
	args := m.Called(ctx, userID, days, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Todo), args.Get(1).(int64), args.Error(2)
}

// Search searches todos by query
func (m *MockTodoRepository) Search(ctx context.Context, userID, query string, limit, offset int) ([]*models.Todo, int64, error) {
	args := m.Called(ctx, userID, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Todo), args.Get(1).(int64), args.Error(2)
}

// CountByStatus counts todos by status
func (m *MockTodoRepository) CountByStatus(ctx context.Context, userID string) (map[string]int64, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

// MarkCompleted marks a todo as completed
func (m *MockTodoRepository) MarkCompleted(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// BulkUpdateStatus updates status for multiple todos
func (m *MockTodoRepository) BulkUpdateStatus(ctx context.Context, ids []string, status string) error {
	args := m.Called(ctx, ids, status)
	return args.Error(0)
}

// DeleteCompleted deletes all completed todos for a user
func (m *MockTodoRepository) DeleteCompleted(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
