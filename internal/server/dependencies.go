package server

import (
	"time"

	"go-fiber/internal/database/mongodb"
	"go-fiber/internal/database/postgres"
	"go-fiber/internal/handlers"
	"go-fiber/internal/repository"
	"go-fiber/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
)

// setupDependencies initializes repositories, services, and handlers
func (s *Server) setupDependencies() error {
	s.logger.Info().Str("driver", s.config.Database.Driver).Msg("Setting up repositories")

	// Determine database type from config
	var dbType repository.DatabaseType
	if s.config.Database.Driver == "postgres" {
		dbType = repository.PostgreSQL
	} else {
		dbType = repository.MongoDB
	}

	// Create repository factory
	repoFactory := repository.NewRepositoryFactory(dbType, s.logger)

	// Setup database connections based on driver
	var pgDB *pgxpool.Pool
	var mongoDB *mongo.Database
	var err error

	if s.config.Database.Driver == "postgres" {
		// Setup PostgreSQL connection
		pgConn, err := postgres.New(&s.config.Database, s.logger)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to connect to PostgreSQL")
			return err
		}
		pgDB = pgConn.Pool
		s.logger.Info().Msg("Successfully connected to PostgreSQL")
	} else {
		// Setup MongoDB connection
		mongoConfig := mongodb.Config{
			URI:      s.config.Database.MongoURL,
			Database: "todoapp", // Extract from URL or make configurable
			Timeout:  10 * time.Second,
		}

		mongoConn, err := mongodb.NewConnection(mongoConfig, s.logger)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to connect to MongoDB")
			return err
		}
		mongoDB = mongoConn.Database
		s.logger.Info().Msg("Successfully connected to MongoDB")
	}

	// Create repositories with actual database connections
	userRepo, err := repoFactory.CreateUserRepository(pgDB, mongoDB)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create user repository")
		return err
	}

	todoRepo, err := repoFactory.CreateTodoRepository(pgDB, mongoDB)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create todo repository")
		return err
	}

	// Setup services
	sessionStore := services.NewRedisSessionStore(s.redisClient, s.logger)
	s.authService = services.NewAuthService(userRepo, sessionStore, &s.config.JWT, s.logger)

	// Setup handlers
	s.authHandler = handlers.NewAuthHandler(s.authService, s.validator, s.logger)
	s.todoHandler = handlers.NewTodoHandler(todoRepo, s.validator, s.logger)
	s.healthHandler = handlers.NewHealthHandler(pgDB, mongoDB, s.redisClient, s.logger)

	s.logger.Info().Msg("Successfully initialized all dependencies")
	return nil
}
