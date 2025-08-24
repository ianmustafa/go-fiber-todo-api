package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config holds MongoDB connection configuration
type Config struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// Connection wraps MongoDB client and database
type Connection struct {
	Client   *mongo.Client
	Database *mongo.Database
	logger   zerolog.Logger
}

// NewConnection creates a new MongoDB connection
func NewConnection(config Config, logger zerolog.Logger) (*Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(config.URI)

	// Create client
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to MongoDB.")
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		logger.Error().Err(err).Msg("Failed to ping MongoDB.")
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	logger.Info().
		Str("database", config.Database).
		Msg("Successfully connected to MongoDB")

	return &Connection{
		Client:   client,
		Database: database,
		logger:   logger,
	}, nil
}

// Close closes the MongoDB connection
func (c *Connection) Close(ctx context.Context) error {
	if err := c.Client.Disconnect(ctx); err != nil {
		c.logger.Error().Err(err).Msg("Failed to disconnect from MongoDB.")
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	c.logger.Info().Msg("Successfully disconnected from MongoDB.")
	return nil
}

// Ping checks if the MongoDB connection is alive
func (c *Connection) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx, readpref.Primary())
}

// GetCollection returns a MongoDB collection
func (c *Connection) GetCollection(name string) *mongo.Collection {
	return c.Database.Collection(name)
}
