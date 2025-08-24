package postgres

import (
	"context"
	"fmt"
	"time"

	"go-fiber/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// DB represents a PostgreSQL database connection
type DB struct {
	Pool   *pgxpool.Pool
	config *config.DatabaseConfig
	logger zerolog.Logger
}

// New creates a new PostgreSQL database connection
func New(cfg *config.DatabaseConfig, logger zerolog.Logger) (*DB, error) {
	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("postgres URL is required")
	}

	// Parse the connection string and configure the pool
	poolConfig, err := pgxpool.ParseConfig(cfg.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres URL: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	poolConfig.HealthCheckPeriod = time.Minute

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	db := &DB{
		Pool:   pool,
		config: cfg,
		logger: logger,
	}

	// Test the connection
	if err := db.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().
		Str("driver", "postgres").
		Int("max_open_conns", cfg.MaxOpenConns).
		Int("max_idle_conns", cfg.MaxIdleConns).
		Msg("PostgreSQL connection established")

	return db, nil
}

// Ping tests the database connection
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.logger.Info().Msg("Closing PostgreSQL connection pool.")
		db.Pool.Close()
	}
}

// Health returns the health status of the database
func (db *DB) Health(ctx context.Context) error {
	// Check if we can ping the database
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check pool stats
	stats := db.Pool.Stat()
	if stats.TotalConns() == 0 {
		return fmt.Errorf("no database connections available")
	}

	return nil
}

// Stats returns connection pool statistics
func (db *DB) Stats() map[string]interface{} {
	stats := db.Pool.Stat()
	return map[string]interface{}{
		"total_conns":        stats.TotalConns(),
		"acquired_conns":     stats.AcquiredConns(),
		"idle_conns":         stats.IdleConns(),
		"constructing_conns": stats.ConstructingConns(),
		"max_conns":          stats.MaxConns(),
		"acquire_count":      stats.AcquireCount(),
		"acquire_duration":   stats.AcquireDuration(),
		"canceled_acquire":   stats.CanceledAcquireCount(),
		"empty_acquire":      stats.EmptyAcquireCount(),
	}
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return db.Pool.Begin(ctx)
}

// WithTx executes a function within a transaction
func (db *DB) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
