package models

import (
	"time"
)

// Todo represents a todo item in the system
type Todo struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"userId" db:"user_id"`
	Title       string     `json:"title" db:"title" validate:"required,min=1,max=200"`
	Description string     `json:"description" db:"description"`
	Status      string     `json:"status" db:"status" validate:"required,oneof=pending in_progress completed"`
	Priority    string     `json:"priority" db:"priority" validate:"oneof=low medium high"`
	DueDate     *time.Time `json:"dueDate,omitempty" db:"due_date"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
}

// GetTodosQueryParams represents query parameters for getting todos
type GetTodosQueryParams struct {
	Limit    int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset   int    `query:"offset" validate:"omitempty,min=0"`
	Status   string `query:"status" validate:"omitempty,oneof=pending in_progress completed"`
	Priority string `query:"priority" validate:"omitempty,oneof=low medium high"`
}

// PaginationQueryParams represents basic pagination query parameters
type PaginationQueryParams struct {
	Limit  int `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

// SearchTodosQueryParams represents query parameters for searching todos
type SearchTodosQueryParams struct {
	Query  string `query:"q" validate:"required,min=1"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `query:"offset" validate:"omitempty,min=0"`
}

// SetDefaults sets default values for query parameters
func (q *GetTodosQueryParams) SetDefaults() {
	if q.Limit == 0 {
		q.Limit = 10
	}
}

// SetDefaults sets default values for pagination parameters
func (p *PaginationQueryParams) SetDefaults() {
	if p.Limit == 0 {
		p.Limit = 10
	}
}

// SetDefaults sets default values for search parameters
func (s *SearchTodosQueryParams) SetDefaults() {
	if s.Limit == 0 {
		s.Limit = 10
	}
}

// CreateTodoRequest represents the request to create a new todo
type CreateTodoRequest struct {
	Title       string     `json:"title" validate:"required,min=1,max=200"`
	Description string     `json:"description,omitempty"`
	Priority    string     `json:"priority,omitempty" validate:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
}

// UpdateTodoRequest represents the request to update a todo
type UpdateTodoRequest struct {
	Title       string     `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress completed"`
	Priority    string     `json:"priority,omitempty" validate:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
}

// UpdateTodoStatusRequest represents the request to update todo status
type UpdateTodoStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending in_progress completed"`
}

// TodoListResponse represents the response for listing todos
type TodoListResponse struct {
	Todos  []*Todo `json:"todos"`
	Total  int64   `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

// TodoStatus constants
const (
	TodoStatusPending    = "pending"
	TodoStatusInProgress = "in_progress"
	TodoStatusCompleted  = "completed"
)

// TodoPriority constants
const (
	TodoPriorityLow    = "low"
	TodoPriorityMedium = "medium"
	TodoPriorityHigh   = "high"
)

// IsValidStatus checks if the status is valid
func IsValidStatus(status string) bool {
	switch status {
	case TodoStatusPending, TodoStatusInProgress, TodoStatusCompleted:
		return true
	default:
		return false
	}
}

// IsValidPriority checks if the priority is valid
func IsValidPriority(priority string) bool {
	switch priority {
	case TodoPriorityLow, TodoPriorityMedium, TodoPriorityHigh:
		return true
	default:
		return false
	}
}

// SetDefaults sets default values for the todo
func (t *Todo) SetDefaults() {
	if t.Status == "" {
		t.Status = TodoStatusPending
	}
	if t.Priority == "" {
		t.Priority = TodoPriorityMedium
	}
}
