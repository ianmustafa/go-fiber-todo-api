package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-fiber/internal/config"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Client wraps the Redis client with additional functionality
type Client struct {
	*redis.Client
	logger zerolog.Logger
	config *config.RedisConfig
}

// NewClient creates a new Redis client with robust URL parsing
func NewClient(cfg *config.RedisConfig, logger zerolog.Logger) (*Client, error) {
	options, err := parseRedisURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Override with explicit config values if provided
	if cfg.Password != "" {
		options.Password = cfg.Password
	}
	if cfg.DB != 0 {
		options.DB = cfg.DB
	}

	// Set connection pool settings
	options.PoolSize = 10
	options.MinIdleConns = 5
	options.MaxIdleConns = 10
	options.ConnMaxIdleTime = 5 * time.Minute
	options.ConnMaxLifetime = 1 * time.Hour

	// Set timeouts
	options.DialTimeout = 5 * time.Second
	options.ReadTimeout = 3 * time.Second
	options.WriteTimeout = 3 * time.Second

	client := redis.NewClient(options)

	redisClient := &Client{
		Client: client,
		logger: logger,
		config: cfg,
	}

	// Test connection
	if err := redisClient.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info().
		Str("addr", options.Addr).
		Int("db", options.DB).
		Msg("Successfully connected to Redis")

	return redisClient, nil
}

// parseRedisURL parses a Redis URL and returns Redis options
func parseRedisURL(redisURL string) (*redis.Options, error) {
	if redisURL == "" {
		// Default configuration
		return &redis.Options{
			Addr: "localhost:6379",
			DB:   0,
		}, nil
	}

	// Handle simple host:port format
	if !strings.Contains(redisURL, "://") {
		return &redis.Options{
			Addr: redisURL,
			DB:   0,
		}, nil
	}

	// Parse full URL
	u, err := url.Parse(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL format: %w", err)
	}

	// Validate scheme
	if u.Scheme != "redis" && u.Scheme != "rediss" {
		return nil, fmt.Errorf("unsupported Redis URL scheme: %s (expected 'redis' or 'rediss')", u.Scheme)
	}

	options := &redis.Options{
		Addr: u.Host,
		DB:   0,
	}

	// Handle TLS for rediss://
	if u.Scheme == "rediss" {
		options.TLSConfig = &tls.Config{
			ServerName: strings.Split(u.Host, ":")[0],
		}
	}

	// Extract password from URL
	if u.User != nil {
		if password, ok := u.User.Password(); ok {
			options.Password = password
		}
	}

	// Extract database number from path
	if u.Path != "" && u.Path != "/" {
		dbStr := strings.TrimPrefix(u.Path, "/")
		if db, err := strconv.Atoi(dbStr); err == nil {
			options.DB = db
		} else {
			return nil, fmt.Errorf("invalid database number in Redis URL: %s", dbStr)
		}
	}

	// Parse query parameters for additional options
	query := u.Query()

	// Connection pool settings
	if poolSize := query.Get("pool_size"); poolSize != "" {
		if size, err := strconv.Atoi(poolSize); err == nil && size > 0 {
			options.PoolSize = size
		}
	}

	if minIdle := query.Get("min_idle_conns"); minIdle != "" {
		if conns, err := strconv.Atoi(minIdle); err == nil && conns >= 0 {
			options.MinIdleConns = conns
		}
	}

	// Timeout settings
	if dialTimeout := query.Get("dial_timeout"); dialTimeout != "" {
		if timeout, err := time.ParseDuration(dialTimeout); err == nil {
			options.DialTimeout = timeout
		}
	}

	if readTimeout := query.Get("read_timeout"); readTimeout != "" {
		if timeout, err := time.ParseDuration(readTimeout); err == nil {
			options.ReadTimeout = timeout
		}
	}

	if writeTimeout := query.Get("write_timeout"); writeTimeout != "" {
		if timeout, err := time.ParseDuration(writeTimeout); err == nil {
			options.WriteTimeout = timeout
		}
	}

	return options, nil
}

// Ping tests the Redis connection
func (c *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Client.Ping(ctx).Err(); err != nil {
		c.logger.Error().Err(err).Msg("Redis ping failed")
		return err
	}

	return nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if err := c.Client.Close(); err != nil {
		c.logger.Error().Err(err).Msg("Failed to close Redis connection")
		return err
	}

	c.logger.Info().Msg("Redis connection closed")
	return nil
}

// GetStats returns Redis connection statistics
func (c *Client) GetStats() *redis.PoolStats {
	return c.Client.PoolStats()
}

// HealthCheck performs a comprehensive health check
func (c *Client) HealthCheck(ctx context.Context) error {
	// Test basic connectivity
	if err := c.Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Test basic operations
	testKey := "health_check_test"
	testValue := "ok"

	// Set a test value
	if err := c.Client.Set(ctx, testKey, testValue, time.Second).Err(); err != nil {
		return fmt.Errorf("set operation failed: %w", err)
	}

	// Get the test value
	result, err := c.Client.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("get operation failed: %w", err)
	}

	if result != testValue {
		return fmt.Errorf("value mismatch: expected %s, got %s", testValue, result)
	}

	// Clean up test key
	c.Client.Del(ctx, testKey)

	return nil
}
