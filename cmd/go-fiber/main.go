package main

import (
	"log"
	"os"

	"go-fiber/internal/config"
	"go-fiber/internal/server"

	"github.com/rs/zerolog"
)

// @title Go Fiber API
// @version 1.0
// @description A production-ready Go API built with Fiber framework
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:9000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Setup logger
	logger := setupLogger(cfg)

	logger.Info().
		Str("environment", cfg.Server.Environment).
		Str("version", "1.0.0").
		Msg("Starting Go Fiber application")

	// Create and start server
	srv := server.New(cfg, logger)
	if err := srv.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server.")
	}
}

// setupLogger configures and returns a zerolog logger
func setupLogger(cfg *config.Config) zerolog.Logger {
	// Set log level
	var level zerolog.Level
	switch cfg.Log.Level {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	// Configure output format
	var logger zerolog.Logger
	if cfg.IsNotProduction() && cfg.Log.Format != "json" {
		// Pretty console output for development
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}).With().Timestamp().Logger()
	} else {
		// JSON output for production
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return logger
}
