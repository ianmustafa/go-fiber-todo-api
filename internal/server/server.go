package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-fiber/internal/config"
	"go-fiber/internal/handlers"
	"go-fiber/internal/services"

	_ "go-fiber/docs" // Import generated docs

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Server represents the HTTP server with all dependencies
type Server struct {
	app         *fiber.App
	config      *config.Config
	logger      zerolog.Logger
	redisClient *redis.Client
	validator   *validator.Validate

	// Services
	authService *services.AuthService

	// Handlers
	authHandler   *handlers.AuthHandler
	todoHandler   *handlers.TodoHandler
	healthHandler *handlers.HealthHandler
}

// New creates a new server instance with all dependencies
func New(cfg *config.Config, logger zerolog.Logger) *Server {
	return &Server{
		config:    cfg,
		logger:    logger,
		validator: validator.New(),
	}
}

// Initialize sets up all dependencies and configurations
func (s *Server) Initialize() error {
	// Setup Fiber app
	s.setupFiberApp()

	// Setup Redis client
	if err := s.setupRedis(); err != nil {
		return err
	}

	// Setup repositories and services
	if err := s.setupDependencies(); err != nil {
		return err
	}

	// Setup middleware
	s.setupMiddleware()

	// Setup routes
	s.setupRoutes()

	return nil
}

// Start starts the server with graceful shutdown
func (s *Server) Start() error {
	// Initialize all dependencies
	if err := s.Initialize(); err != nil {
		return err
	}

	// Start server in a goroutine
	go func() {
		address := s.config.GetAddress()
		s.logger.Info().
			Str("address", address).
			Str("environment", s.config.Server.Environment).
			Msg("Starting server.")

		if err := s.app.Listen(address); err != nil {
			s.logger.Fatal().Err(err).Msg("Failed to start server.")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Server forced to shutdown.")
		return err
	}

	// Close Redis connection
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			s.logger.Error().Err(err).Msg("Failed to close Redis connection.")
		}
	}

	s.logger.Info().Msg("Server exited.")
	return nil
}

// GetApp returns the Fiber app instance for testing
func (s *Server) GetApp() *fiber.App {
	return s.app
}
