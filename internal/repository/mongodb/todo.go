package mongodb

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"go-fiber/internal/models"
	"go-fiber/internal/repository/interfaces"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoTodo represents a todo document in MongoDB
type MongoTodo struct {
	ID          string     `bson:"_id" json:"id"`
	UserID      string     `bson:"userId" json:"userId"`
	Title       string     `bson:"title" json:"title"`
	Description string     `bson:"description,omitempty" json:"description,omitempty"`
	Status      string     `bson:"status" json:"status"`
	Priority    string     `bson:"priority,omitempty" json:"priority,omitempty"`
	DueDate     *time.Time `bson:"dueDate,omitempty" json:"dueDate,omitempty"`
	CreatedAt   time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time  `bson:"updatedAt" json:"updatedAt"`
	DeletedAt   *time.Time `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
}

// todoRepository implements the TodoRepository interface for MongoDB
type todoRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewTodoRepository creates a new MongoDB todo repository
func NewTodoRepository(db *mongo.Database, logger zerolog.Logger) interfaces.TodoRepository {
	return &todoRepository{
		collection: db.Collection("todos"),
		logger:     logger,
	}
}

// Create creates a new todo
func (r *todoRepository) Create(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
	// Generate ULID for new todo
	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)

	now := time.Now()

	// Set defaults
	status := todo.Status
	if status == "" {
		status = models.TodoStatusPending
	}

	priority := todo.Priority
	if priority == "" {
		priority = models.TodoPriorityMedium
	}

	mongoTodo := &MongoTodo{
		ID:          id.String(),
		UserID:      todo.UserID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      status,
		Priority:    priority,
		DueDate:     todo.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := r.collection.InsertOne(ctx, mongoTodo)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", todo.UserID).Str("title", todo.Title).Msg("Failed to create todo.")
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	result := r.mongoTodoToModel(mongoTodo)
	r.logger.Info().Str("todo_id", result.ID).Str("user_id", result.UserID).Msg("Todo created successfully.")
	return result, nil
}

// GetByID retrieves a todo by ID
func (r *todoRepository) GetByID(ctx context.Context, id string) (*models.Todo, error) {
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}

	var mongoTodo MongoTodo
	err := r.collection.FindOne(ctx, filter).Decode(&mongoTodo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("todo not found")
		}
		r.logger.Error().Err(err).Str("todo_id", id).Msg("Failed to get todo by ID.")
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return r.mongoTodoToModel(&mongoTodo), nil
}

// GetByUserID retrieves todos by user ID with pagination
func (r *todoRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Todo, int64, error) {
	filter := bson.M{
		"userId":    userID,
		"deletedAt": bson.M{"$exists": false},
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to count todos by user ID.")
		return nil, 0, fmt.Errorf("failed to count todos: %w", err)
	}

	// Get todos with pagination
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get todos by user ID.")
		return nil, 0, fmt.Errorf("failed to get todos: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoTodos []MongoTodo
	if err := cursor.All(ctx, &mongoTodos); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode todos.")
		return nil, 0, fmt.Errorf("failed to decode todos: %w", err)
	}

	todos := make([]*models.Todo, len(mongoTodos))
	for i, mongoTodo := range mongoTodos {
		todos[i] = r.mongoTodoToModel(&mongoTodo)
	}

	return todos, total, nil
}

// Update updates a todo
func (r *todoRepository) Update(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
	filter := bson.M{
		"_id":       todo.ID,
		"deletedAt": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"title":       todo.Title,
			"description": todo.Description,
			"status":      todo.Status,
			"priority":    todo.Priority,
			"dueDate":     todo.DueDate,
			"updatedAt":   time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var mongoTodo MongoTodo
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&mongoTodo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("todo not found")
		}
		r.logger.Error().Err(err).Str("todo_id", todo.ID).Msg("Failed to update todo.")
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	result := r.mongoTodoToModel(&mongoTodo)
	r.logger.Info().Str("todo_id", result.ID).Msg("Todo updated successfully.")
	return result, nil
}

// Delete soft deletes a todo
func (r *todoRepository) Delete(ctx context.Context, id string) error {
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"deletedAt": time.Now(),
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("todo_id", id).Msg("Failed to delete todo.")
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("todo not found")
	}

	r.logger.Info().Str("todo_id", id).Msg("Todo deleted successfully.")
	return nil
}

// UpdateStatus updates a todo's status
func (r *todoRepository) UpdateStatus(ctx context.Context, id, status string) error {
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("todo_id", id).Str("status", status).Msg("Failed to update todo status.")
		return fmt.Errorf("failed to update todo status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("todo not found")
	}

	r.logger.Info().Str("todo_id", id).Str("status", status).Msg("Todo status updated successfully.")
	return nil
}

// GetByStatus retrieves todos by status with pagination
func (r *todoRepository) GetByStatus(ctx context.Context, userID, status string, limit, offset int) ([]*models.Todo, int64, error) {
	filter := bson.M{
		"userId":    userID,
		"status":    status,
		"deletedAt": bson.M{"$exists": false},
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("status", status).Msg("Failed to count todos by status.")
		return nil, 0, fmt.Errorf("failed to count todos: %w", err)
	}

	// Get todos with pagination
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("status", status).Msg("Failed to get todos by status.")
		return nil, 0, fmt.Errorf("failed to get todos: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoTodos []MongoTodo
	if err := cursor.All(ctx, &mongoTodos); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode todos.")
		return nil, 0, fmt.Errorf("failed to decode todos: %w", err)
	}

	todos := make([]*models.Todo, len(mongoTodos))
	for i, mongoTodo := range mongoTodos {
		todos[i] = r.mongoTodoToModel(&mongoTodo)
	}

	return todos, total, nil
}

// GetByPriority retrieves todos by priority with pagination
func (r *todoRepository) GetByPriority(ctx context.Context, userID, priority string, limit, offset int) ([]*models.Todo, int64, error) {
	filter := bson.M{
		"userId":    userID,
		"priority":  priority,
		"deletedAt": bson.M{"$exists": false},
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("priority", priority).Msg("Failed to count todos by priority.")
		return nil, 0, fmt.Errorf("failed to count todos: %w", err)
	}

	// Get todos with pagination
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("priority", priority).Msg("Failed to get todos by priority.")
		return nil, 0, fmt.Errorf("failed to get todos: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoTodos []MongoTodo
	if err := cursor.All(ctx, &mongoTodos); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode todos.")
		return nil, 0, fmt.Errorf("failed to decode todos: %w", err)
	}

	todos := make([]*models.Todo, len(mongoTodos))
	for i, mongoTodo := range mongoTodos {
		todos[i] = r.mongoTodoToModel(&mongoTodo)
	}

	return todos, total, nil
}

// GetOverdue retrieves overdue todos with pagination
func (r *todoRepository) GetOverdue(ctx context.Context, userID string, limit, offset int) ([]*models.Todo, int64, error) {
	now := time.Now()
	filter := bson.M{
		"userId":    userID,
		"dueDate":   bson.M{"$lt": now},
		"status":    bson.M{"$ne": models.TodoStatusCompleted},
		"deletedAt": bson.M{"$exists": false},
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to count overdue todos.")
		return nil, 0, fmt.Errorf("failed to count overdue todos: %w", err)
	}

	// Get todos with pagination
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"dueDate": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get overdue todos.")
		return nil, 0, fmt.Errorf("failed to get overdue todos: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoTodos []MongoTodo
	if err := cursor.All(ctx, &mongoTodos); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode todos.")
		return nil, 0, fmt.Errorf("failed to decode todos: %w", err)
	}

	todos := make([]*models.Todo, len(mongoTodos))
	for i, mongoTodo := range mongoTodos {
		todos[i] = r.mongoTodoToModel(&mongoTodo)
	}

	return todos, total, nil
}

// GetUpcoming retrieves upcoming todos with pagination
func (r *todoRepository) GetUpcoming(ctx context.Context, userID string, days int, limit, offset int) ([]*models.Todo, int64, error) {
	now := time.Now()
	futureDate := now.AddDate(0, 0, days)

	filter := bson.M{
		"userId": userID,
		"dueDate": bson.M{
			"$gte": now,
			"$lte": futureDate,
		},
		"status":    bson.M{"$ne": models.TodoStatusCompleted},
		"deletedAt": bson.M{"$exists": false},
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to count upcoming todos.")
		return nil, 0, fmt.Errorf("failed to count upcoming todos: %w", err)
	}

	// Get todos with pagination
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"dueDate": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get upcoming todos.")
		return nil, 0, fmt.Errorf("failed to get upcoming todos: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoTodos []MongoTodo
	if err := cursor.All(ctx, &mongoTodos); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode todos.")
		return nil, 0, fmt.Errorf("failed to decode todos: %w", err)
	}

	todos := make([]*models.Todo, len(mongoTodos))
	for i, mongoTodo := range mongoTodos {
		todos[i] = r.mongoTodoToModel(&mongoTodo)
	}

	return todos, total, nil
}

// Search searches todos with pagination
func (r *todoRepository) Search(ctx context.Context, userID, query string, limit, offset int) ([]*models.Todo, int64, error) {
	filter := bson.M{
		"userId":    userID,
		"deletedAt": bson.M{"$exists": false},
		"$text":     bson.M{"$search": query},
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("query", query).Msg("Failed to count search todos.")
		return nil, 0, fmt.Errorf("failed to count search todos: %w", err)
	}

	// Get todos with pagination
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"score": bson.M{"$meta": "textScore"}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("query", query).Msg("Failed to search todos.")
		return nil, 0, fmt.Errorf("failed to search todos: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoTodos []MongoTodo
	if err := cursor.All(ctx, &mongoTodos); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode todos.")
		return nil, 0, fmt.Errorf("failed to decode todos: %w", err)
	}

	todos := make([]*models.Todo, len(mongoTodos))
	for i, mongoTodo := range mongoTodos {
		todos[i] = r.mongoTodoToModel(&mongoTodo)
	}

	return todos, total, nil
}

// CountByStatus returns count of todos by status
func (r *todoRepository) CountByStatus(ctx context.Context, userID string) (map[string]int64, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"userId":    userID,
				"deletedAt": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$status",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get todo status counts.")
		return nil, fmt.Errorf("failed to get todo status counts: %w", err)
	}
	defer cursor.Close(ctx)

	counts := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			Status string `bson:"_id"`
			Count  int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			r.logger.Error().Err(err).Msg("Failed to decode status count.")
			continue
		}
		counts[result.Status] = result.Count
	}

	return counts, nil
}

// MarkCompleted marks a todo as completed
func (r *todoRepository) MarkCompleted(ctx context.Context, id string) error {
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"status":    models.TodoStatusCompleted,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("todo_id", id).Msg("Failed to mark todo as completed.")
		return fmt.Errorf("failed to mark todo as completed: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("todo not found")
	}

	r.logger.Info().Str("todo_id", id).Msg("Todo marked as completed.")
	return nil
}

// BulkUpdateStatus updates status for multiple todos
func (r *todoRepository) BulkUpdateStatus(ctx context.Context, ids []string, status string) error {
	filter := bson.M{
		"_id":       bson.M{"$in": ids},
		"deletedAt": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Strs("todo_ids", ids).Str("status", status).Msg("Failed to bulk update todo status.")
		return fmt.Errorf("failed to bulk update todo status: %w", err)
	}

	r.logger.Info().Strs("todo_ids", ids).Str("status", status).Int64("updated_count", result.ModifiedCount).Msg("Todos status updated in bulk.")
	return nil
}

// DeleteCompleted soft deletes all completed todos for a user
func (r *todoRepository) DeleteCompleted(ctx context.Context, userID string) error {
	filter := bson.M{
		"userId":    userID,
		"status":    models.TodoStatusCompleted,
		"deletedAt": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"deletedAt": time.Now(),
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to delete completed todos.")
		return fmt.Errorf("failed to delete completed todos: %w", err)
	}

	r.logger.Info().Str("user_id", userID).Int64("deleted_count", result.ModifiedCount).Msg("Completed todos deleted.")
	return nil
}

// mongoTodoToModel converts a MongoDB todo document to a model todo
func (r *todoRepository) mongoTodoToModel(mongoTodo *MongoTodo) *models.Todo {
	return &models.Todo{
		ID:          mongoTodo.ID,
		UserID:      mongoTodo.UserID,
		Title:       mongoTodo.Title,
		Description: mongoTodo.Description,
		Status:      mongoTodo.Status,
		Priority:    mongoTodo.Priority,
		DueDate:     mongoTodo.DueDate,
		CreatedAt:   mongoTodo.CreatedAt,
		UpdatedAt:   mongoTodo.UpdatedAt,
	}
}
