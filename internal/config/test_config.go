package config

import (
	"time"

	"github.com/rs/zerolog"
)

// NewTestConfig creates a configuration suitable for testing
func NewTestConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "localhost",
			Port:         9000,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Environment:  "test",
		},
		Database: DatabaseConfig{
			Driver:       "postgres",
			PostgresURL:  "postgres://test:test@localhost:5432/test_db?sslmode=disable",
			MongoURL:     "mongodb://localhost:27017/test_db",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		Redis: RedisConfig{
			URL:      "redis://localhost:6379/1", // Use DB 1 for tests
			Password: "",
			DB:       1,
		},
		JWT: JWTConfig{
			Secret:        "test-secret-key-for-testing-only-must-be-32-chars",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 24 * time.Hour,
			Issuer:        "go-fiber-test",
		},
		Log: LogConfig{
			Level:  "debug",
			Format: "json",
		},
		RateLimit: RateLimitConfig{
			Requests: 1000, // High limit for tests
			Window:   time.Minute,
		},
	}
}

// NewTestLogger creates a logger suitable for testing
func NewTestLogger() zerolog.Logger {
	return zerolog.Nop() // No-op logger for tests
}
