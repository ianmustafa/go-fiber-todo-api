package interfaces

import (
	"context"

	"go-fiber/internal/models"
)

// TodoRepository defines the interface for todo data operations
type TodoRepository interface {
	Create(ctx context.Context, todo *models.Todo) (*models.Todo, error)
	GetByID(ctx context.Context, id string) (*models.Todo, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Todo, int64, error)
	Update(ctx context.Context, todo *models.Todo) (*models.Todo, error)
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, status string) error
	GetByStatus(ctx context.Context, userID, status string, limit, offset int) ([]*models.Todo, int64, error)
	GetByPriority(ctx context.Context, userID, priority string, limit, offset int) ([]*models.Todo, int64, error)
	GetOverdue(ctx context.Context, userID string, limit, offset int) ([]*models.Todo, int64, error)
	GetUpcoming(ctx context.Context, userID string, days int, limit, offset int) ([]*models.Todo, int64, error)
	Search(ctx context.Context, userID, query string, limit, offset int) ([]*models.Todo, int64, error)
	CountByStatus(ctx context.Context, userID string) (map[string]int64, error)
	MarkCompleted(ctx context.Context, id string) error
	BulkUpdateStatus(ctx context.Context, ids []string, status string) error
	DeleteCompleted(ctx context.Context, userID string) error
}
