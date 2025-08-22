package server

import (
	redisDB "go-fiber/internal/database/redis"
)

// setupRedis initializes Redis client using the database package
func (s *Server) setupRedis() error {
	client, err := redisDB.NewClient(&s.config.Redis, s.logger)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create Redis client")
		return err
	}

	// Store the underlying Redis client for compatibility
	s.redisClient = client.Client

	s.logger.Info().Msg("Redis client setup completed")
	return nil
}
