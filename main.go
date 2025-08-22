// Package main Go Fiber Todo API
//
// A comprehensive Todo API built with Go Fiber framework, featuring clean architecture,
// JWT authentication, and multiple database support (PostgreSQL and MongoDB).
//
// @title Go Fiber Todo API
// @version 1.0
// @description A comprehensive Todo API built with Go Fiber framework, featuring clean architecture, JWT authentication, and multiple database support.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name MIT
// @license.url http://opensource.org/licenses/MIT
// @host localhost:9000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"log"
	"os"
	"time"

	"go-fiber/internal/config"
	"go-fiber/internal/server"

	"github.com/rs/zerolog"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Setup logger
	appLogger := setupLogger(cfg)

	// Create and start server
	srv := server.New(cfg, appLogger)
	if err := srv.Start(); err != nil {
		appLogger.Fatal().Err(err).Msg("Server failed to start")
	}
}

// setupLogger configures and returns a structured logger
func setupLogger(cfg *config.Config) zerolog.Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure logger output
	var appLogger zerolog.Logger
	if cfg.Server.Environment != "production" {
		appLogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			With().
			Timestamp().
			Caller().
			Logger()
	} else {
		appLogger = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Caller().
			Logger()
	}

	return appLogger
}
