package mocks

import (
	"context"
	"time"

	"go-fiber/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockSessionStore is a mock implementation of SessionStore
type MockSessionStore struct {
	mock.Mock
}

// Set mocks the Set method
func (m *MockSessionStore) Set(ctx context.Context, sessionID string, session *models.Session, expiration time.Duration) error {
	args := m.Called(ctx, sessionID, session, expiration)
	return args.Error(0)
}

// Get mocks the Get method
func (m *MockSessionStore) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

// Delete mocks the Delete method
func (m *MockSessionStore) Delete(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

// DeleteUserSessions mocks the DeleteUserSessions method
func (m *MockSessionStore) DeleteUserSessions(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
