package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-fiber/internal/models"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// RedisSessionStore implements SessionStore using Redis
type RedisSessionStore struct {
	client redis.Cmdable
	logger zerolog.Logger
	prefix string
}

// NewRedisSessionStore creates a new Redis session store
func NewRedisSessionStore(client redis.Cmdable, logger zerolog.Logger) *RedisSessionStore {
	return &RedisSessionStore{
		client: client,
		logger: logger,
		prefix: "session:",
	}
}

// Set stores a session in Redis
func (s *RedisSessionStore) Set(ctx context.Context, sessionID string, session *models.Session, expiration time.Duration) error {
	key := s.getKey(sessionID)

	// Serialize session to JSON
	data, err := json.Marshal(session)
	if err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to marshal session.")
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Store in Redis with expiration
	if err := s.client.Set(ctx, key, data, expiration).Err(); err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to store session in Redis.")
		return fmt.Errorf("failed to store session: %w", err)
	}

	s.logger.Debug().Str("session_id", sessionID).Dur("expiration", expiration).Msg("Session stored successfully.")
	return nil
}

// Get retrieves a session from Redis
func (s *RedisSessionStore) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	key := s.getKey(sessionID)

	// Get from Redis
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get session from Redis.")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Deserialize session from JSON
	var session models.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to unmarshal session.")
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	s.logger.Debug().Str("session_id", sessionID).Msg("Session retrieved successfully.")
	return &session, nil
}

// Delete removes a session from Redis
func (s *RedisSessionStore) Delete(ctx context.Context, sessionID string) error {
	key := s.getKey(sessionID)

	// Delete from Redis
	result, err := s.client.Del(ctx, key).Result()
	if err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to delete session from Redis.")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result == 0 {
		s.logger.Warn().Str("session_id", sessionID).Msg("Session not found for deletion.")
		return fmt.Errorf("session not found")
	}

	s.logger.Debug().Str("session_id", sessionID).Msg("Session deleted successfully.")
	return nil
}

// DeleteUserSessions removes all sessions for a specific user
func (s *RedisSessionStore) DeleteUserSessions(ctx context.Context, userID string) error {
	// Get all session keys
	pattern := s.prefix + "*"
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get session keys.")
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	// Check each session to see if it belongs to the user
	var userSessionKeys []string
	for _, key := range keys {
		data, err := s.client.Get(ctx, key).Result()
		if err != nil {
			continue // Skip if we can't get the session
		}

		var session models.Session
		if err := json.Unmarshal([]byte(data), &session); err != nil {
			continue // Skip if we can't unmarshal the session
		}

		if session.UserID == userID {
			userSessionKeys = append(userSessionKeys, key)
		}
	}

	// Delete user sessions
	if len(userSessionKeys) > 0 {
		deleted, err := s.client.Del(ctx, userSessionKeys...).Result()
		if err != nil {
			s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user sessions.")
			return fmt.Errorf("failed to delete user sessions: %w", err)
		}

		s.logger.Info().Str("user_id", userID).Int64("deleted_count", deleted).Msg("User sessions deleted successfully.")
	}

	return nil
}

// Exists checks if a session exists in Redis
func (s *RedisSessionStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := s.getKey(sessionID)

	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to check session existence.")
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return result > 0, nil
}

// Extend extends the expiration time of a session
func (s *RedisSessionStore) Extend(ctx context.Context, sessionID string, expiration time.Duration) error {
	key := s.getKey(sessionID)

	// Check if session exists
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to check session existence.")
		return fmt.Errorf("failed to check session existence: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("session not found")
	}

	// Extend expiration
	if err := s.client.Expire(ctx, key, expiration).Err(); err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to extend session expiration.")
		return fmt.Errorf("failed to extend session expiration: %w", err)
	}

	s.logger.Debug().Str("session_id", sessionID).Dur("expiration", expiration).Msg("Session expiration extended.")
	return nil
}

// GetTTL returns the remaining time to live for a session
func (s *RedisSessionStore) GetTTL(ctx context.Context, sessionID string) (time.Duration, error) {
	key := s.getKey(sessionID)

	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get session TTL.")
		return 0, fmt.Errorf("failed to get session TTL: %w", err)
	}

	return ttl, nil
}

// Count returns the total number of active sessions
func (s *RedisSessionStore) Count(ctx context.Context) (int64, error) {
	pattern := s.prefix + "*"
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to count sessions.")
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	return int64(len(keys)), nil
}

// CountUserSessions returns the number of active sessions for a specific user
func (s *RedisSessionStore) CountUserSessions(ctx context.Context, userID string) (int64, error) {
	pattern := s.prefix + "*"
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get session keys.")
		return 0, fmt.Errorf("failed to get session keys: %w", err)
	}

	var count int64
	for _, key := range keys {
		data, err := s.client.Get(ctx, key).Result()
		if err != nil {
			continue // Skip if we can't get the session
		}

		var session models.Session
		if err := json.Unmarshal([]byte(data), &session); err != nil {
			continue // Skip if we can't unmarshal the session
		}

		if session.UserID == userID {
			count++
		}
	}

	return count, nil
}

// Cleanup removes expired sessions (Redis handles this automatically, but this can be used for manual cleanup)
func (s *RedisSessionStore) Cleanup(ctx context.Context) error {
	// Redis automatically handles expiration, but we can implement manual cleanup if needed
	s.logger.Info().Msg("Session cleanup completed (Redis handles expiration automatically).")
	return nil
}

// getKey generates the Redis key for a session
func (s *RedisSessionStore) getKey(sessionID string) string {
	return s.prefix + sessionID
}

// SetPrefix sets the key prefix for sessions
func (s *RedisSessionStore) SetPrefix(prefix string) {
	s.prefix = prefix
}

// GetPrefix returns the current key prefix
func (s *RedisSessionStore) GetPrefix() string {
	return s.prefix
}
