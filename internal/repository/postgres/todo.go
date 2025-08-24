package postgres

import (
	"context"
	"fmt"

	"go-fiber/internal/models"
	"go-fiber/internal/repository/interfaces"
	"go-fiber/internal/repository/postgres/queries"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// todoRepository implements the TodoRepository interface for PostgreSQL
type todoRepository struct {
	db      *pgxpool.Pool
	queries *queries.Queries
	logger  zerolog.Logger
}

// NewTodoRepository creates a new PostgreSQL todo repository
func NewTodoRepository(db *pgxpool.Pool, logger zerolog.Logger) interfaces.TodoRepository {
	return &todoRepository{
		db:      db,
		queries: queries.New(db),
		logger:  logger,
	}
}

// Create creates a new todo
func (r *todoRepository) Create(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
	var description, priority pgtype.Text
	var dueDate pgtype.Timestamptz

	if todo.Description != "" {
		description = pgtype.Text{String: todo.Description, Valid: true}
	}
	if todo.Priority != "" {
		priority = pgtype.Text{String: todo.Priority, Valid: true}
	} else {
		priority = pgtype.Text{String: models.TodoPriorityMedium, Valid: true}
	}
	if todo.DueDate != nil {
		dueDate = pgtype.Timestamptz{Time: *todo.DueDate, Valid: true}
	}

	// Set default status if not provided
	status := todo.Status
	if status == "" {
		status = models.TodoStatusPending
	}

	dbTodo, err := r.queries.CreateTodo(ctx, queries.CreateTodoParams{
		UserID:      todo.UserID,
		Title:       todo.Title,
		Description: description,
		Status:      status,
		Priority:    priority,
		DueDate:     dueDate,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", todo.UserID).Str("title", todo.Title).Msg("Failed to create todo.")
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	result := r.mapDBTodoToModel(dbTodo)
	r.logger.Info().Str("todo_id", result.ID).Str("user_id", result.UserID).Msg("Todo created successfully.")
	return result, nil
}

// GetByID retrieves a todo by ID
func (r *todoRepository) GetByID(ctx context.Context, id string) (*models.Todo, error) {
	dbTodo, err := r.queries.GetTodoByID(ctx, id)
	if err != nil {
		r.logger.Error().Err(err).Str("todo_id", id).Msg("Failed to get todo by ID.")
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return r.mapDBTodoToModel(dbTodo), nil
}

// GetByUserID retrieves todos by user ID with pagination
func (r *todoRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Todo, int64, error) {
	// Get total count
	total, err := r.queries.CountTodosByUserID(ctx, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to count todos by user ID.")
		return nil, 0, fmt.Errorf("failed to count todos: %w", err)
	}

	// Get todos
	dbTodos, err := r.queries.GetTodosByUserID(ctx, queries.GetTodosByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get todos by user ID.")
		return nil, 0, fmt.Errorf("failed to get todos: %w", err)
	}

	todos := make([]*models.Todo, len(dbTodos))
	for i, dbTodo := range dbTodos {
		todos[i] = r.mapDBTodoToModel(dbTodo)
	}

	return todos, total, nil
}

// Update updates a todo
func (r *todoRepository) Update(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
	var description, priority pgtype.Text
	var dueDate pgtype.Timestamptz

	if todo.Description != "" {
		description = pgtype.Text{String: todo.Description, Valid: true}
	}
	if todo.Priority != "" {
		priority = pgtype.Text{String: todo.Priority, Valid: true}
	}
	if todo.DueDate != nil {
		dueDate = pgtype.Timestamptz{Time: *todo.DueDate, Valid: true}
	}

	dbTodo, err := r.queries.UpdateTodo(ctx, queries.UpdateTodoParams{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: description,
		Status:      todo.Status,
		Priority:    priority,
		DueDate:     dueDate,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("todo_id", todo.ID).Msg("Failed to update todo.")
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	result := r.mapDBTodoToModel(dbTodo)
	r.logger.Info().Str("todo_id", result.ID).Msg("Todo updated successfully.")
	return result, nil
}

// Delete soft deletes a todo
func (r *todoRepository) Delete(ctx context.Context, id string) error {
	err := r.queries.SoftDeleteTodo(ctx, id)
	if err != nil {
		r.logger.Error().Err(err).Str("todo_id", id).Msg("Failed to delete todo.")
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	r.logger.Info().Str("todo_id", id).Msg("Todo deleted successfully.")
	return nil
}

// UpdateStatus updates a todo's status
func (r *todoRepository) UpdateStatus(ctx context.Context, id, status string) error {
	err := r.queries.UpdateTodoStatus(ctx, queries.UpdateTodoStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("todo_id", id).Str("status", status).Msg("Failed to update todo status.")
		return fmt.Errorf("failed to update todo status: %w", err)
	}

	r.logger.Info().Str("todo_id", id).Str("status", status).Msg("Todo status updated successfully.")
	return nil
}

// GetByStatus retrieves todos by status with pagination
func (r *todoRepository) GetByStatus(ctx context.Context, userID, status string, limit, offset int) ([]*models.Todo, int64, error) {
	// Get total count
	total, err := r.queries.CountTodosByStatus(ctx, queries.CountTodosByStatusParams{
		UserID: userID,
		Status: status,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("status", status).Msg("Failed to count todos by status.")
		return nil, 0, fmt.Errorf("failed to count todos: %w", err)
	}

	// Get todos
	dbTodos, err := r.queries.GetTodosByStatus(ctx, queries.GetTodosByStatusParams{
		UserID: userID,
		Status: status,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("status", status).Msg("Failed to get todos by status.")
		return nil, 0, fmt.Errorf("failed to get todos: %w", err)
	}

	todos := make([]*models.Todo, len(dbTodos))
	for i, dbTodo := range dbTodos {
		todos[i] = r.mapDBTodoToModel(dbTodo)
	}

	return todos, total, nil
}

// GetByPriority retrieves todos by priority with pagination
func (r *todoRepository) GetByPriority(ctx context.Context, userID, priority string, limit, offset int) ([]*models.Todo, int64, error) {
	// Get total count
	total, err := r.queries.CountTodosByPriority(ctx, queries.CountTodosByPriorityParams{
		UserID:   userID,
		Priority: pgtype.Text{String: priority, Valid: true},
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("priority", priority).Msg("Failed to count todos by priority.")
		return nil, 0, fmt.Errorf("failed to count todos: %w", err)
	}

	// Get todos
	dbTodos, err := r.queries.GetTodosByPriority(ctx, queries.GetTodosByPriorityParams{
		UserID:   userID,
		Priority: pgtype.Text{String: priority, Valid: true},
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("priority", priority).Msg("Failed to get todos by priority.")
		return nil, 0, fmt.Errorf("failed to get todos: %w", err)
	}

	todos := make([]*models.Todo, len(dbTodos))
	for i, dbTodo := range dbTodos {
		todos[i] = r.mapDBTodoToModel(dbTodo)
	}

	return todos, total, nil
}

// GetOverdue retrieves overdue todos with pagination
func (r *todoRepository) GetOverdue(ctx context.Context, userID string, limit, offset int) ([]*models.Todo, int64, error) {
	// Get total count
	total, err := r.queries.CountOverdueTodos(ctx, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to count overdue todos.")
		return nil, 0, fmt.Errorf("failed to count overdue todos: %w", err)
	}

	// Get todos
	dbTodos, err := r.queries.GetOverdueTodos(ctx, queries.GetOverdueTodosParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get overdue todos.")
		return nil, 0, fmt.Errorf("failed to get overdue todos: %w", err)
	}

	todos := make([]*models.Todo, len(dbTodos))
	for i, dbTodo := range dbTodos {
		todos[i] = r.mapDBTodoToModel(dbTodo)
	}

	return todos, total, nil
}

// GetUpcoming retrieves upcoming todos with pagination
func (r *todoRepository) GetUpcoming(ctx context.Context, userID string, days int, limit, offset int) ([]*models.Todo, int64, error) {
	// Note: The SQLC queries need to be updated to handle dynamic intervals
	// For now, we'll implement a basic version
	dbTodos, err := r.queries.GetUpcomingTodos(ctx, queries.GetUpcomingTodosParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get upcoming todos.")
		return nil, 0, fmt.Errorf("failed to get upcoming todos: %w", err)
	}

	// Get count
	total, err := r.queries.CountUpcomingTodos(ctx, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to count upcoming todos.")
		return nil, 0, fmt.Errorf("failed to count upcoming todos: %w", err)
	}

	todos := make([]*models.Todo, len(dbTodos))
	for i, dbTodo := range dbTodos {
		todos[i] = r.mapDBTodoToModel(dbTodo)
	}

	return todos, total, nil
}

// Search searches todos with pagination
func (r *todoRepository) Search(ctx context.Context, userID, query string, limit, offset int) ([]*models.Todo, int64, error) {
	// Get total count
	total, err := r.queries.CountSearchTodos(ctx, queries.CountSearchTodosParams{
		UserID:         userID,
		PlaintoTsquery: query,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("query", query).Msg("Failed to count search todos.")
		return nil, 0, fmt.Errorf("failed to count search todos: %w", err)
	}

	// Get todos
	dbTodos, err := r.queries.SearchTodos(ctx, queries.SearchTodosParams{
		UserID:         userID,
		PlaintoTsquery: query,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("query", query).Msg("Failed to search todos.")
		return nil, 0, fmt.Errorf("failed to search todos: %w", err)
	}

	todos := make([]*models.Todo, len(dbTodos))
	for i, dbTodo := range dbTodos {
		todos[i] = r.mapDBTodoToModel(dbTodo)
	}

	return todos, total, nil
}

// CountByStatus returns count of todos by status
func (r *todoRepository) CountByStatus(ctx context.Context, userID string) (map[string]int64, error) {
	rows, err := r.queries.GetTodoStatusCounts(ctx, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get todo status counts.")
		return nil, fmt.Errorf("failed to get todo status counts: %w", err)
	}

	counts := make(map[string]int64)
	for _, row := range rows {
		counts[row.Status] = row.Count
	}

	return counts, nil
}

// MarkCompleted marks a todo as completed
func (r *todoRepository) MarkCompleted(ctx context.Context, id string) error {
	err := r.queries.MarkTodoCompleted(ctx, id)
	if err != nil {
		r.logger.Error().Err(err).Str("todo_id", id).Msg("Failed to mark todo as completed.")
		return fmt.Errorf("failed to mark todo as completed: %w", err)
	}

	r.logger.Info().Str("todo_id", id).Msg("Todo marked as completed.")
	return nil
}

// BulkUpdateStatus updates status for multiple todos
func (r *todoRepository) BulkUpdateStatus(ctx context.Context, ids []string, status string) error {
	// Convert []string to []interface{}
	interfaceIds := make([]interface{}, len(ids))
	for i, id := range ids {
		interfaceIds[i] = id
	}

	err := r.queries.BulkUpdateTodoStatus(ctx, queries.BulkUpdateTodoStatusParams{
		Column1: interfaceIds,
		Status:  status,
	})
	if err != nil {
		r.logger.Error().Err(err).Strs("todo_ids", ids).Str("status", status).Msg("Failed to bulk update todo status.")
		return fmt.Errorf("failed to bulk update todo status: %w", err)
	}

	r.logger.Info().Strs("todo_ids", ids).Str("status", status).Msg("Todos status updated in bulk.")
	return nil
}

// DeleteCompleted soft deletes all completed todos for a user
func (r *todoRepository) DeleteCompleted(ctx context.Context, userID string) error {
	err := r.queries.SoftDeleteCompletedTodos(ctx, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to delete completed todos.")
		return fmt.Errorf("failed to delete completed todos: %w", err)
	}

	r.logger.Info().Str("user_id", userID).Msg("Completed todos deleted.")
	return nil
}

// mapDBTodoToModel converts a database todo to a model todo
func (r *todoRepository) mapDBTodoToModel(dbTodo queries.Todo) *models.Todo {
	todo := &models.Todo{
		ID:        fmt.Sprintf("%v", dbTodo.ID),     // Convert interface{} to string
		UserID:    fmt.Sprintf("%v", dbTodo.UserID), // Convert interface{} to string
		Title:     dbTodo.Title,
		Status:    dbTodo.Status,
		CreatedAt: dbTodo.CreatedAt.Time,
		UpdatedAt: dbTodo.UpdatedAt.Time,
	}

	if dbTodo.Description.Valid {
		todo.Description = dbTodo.Description.String
	}
	if dbTodo.Priority.Valid {
		todo.Priority = dbTodo.Priority.String
	}
	if dbTodo.DueDate.Valid {
		todo.DueDate = &dbTodo.DueDate.Time
	}

	return todo
}
