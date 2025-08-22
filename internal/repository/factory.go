package repository

import (
	"fmt"

	"go-fiber/internal/repository/interfaces"
	mongoRepo "go-fiber/internal/repository/mongodb"
	postgresRepo "go-fiber/internal/repository/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
	MongoDB    DatabaseType = "mongodb"
)

// RepositoryFactory creates repository instances based on database type
type RepositoryFactory struct {
	dbType DatabaseType
	logger zerolog.Logger
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(dbType DatabaseType, logger zerolog.Logger) *RepositoryFactory {
	return &RepositoryFactory{
		dbType: dbType,
		logger: logger,
	}
}

// CreateUserRepository creates a user repository based on database type
func (f *RepositoryFactory) CreateUserRepository(pgDB *pgxpool.Pool, mongoDB *mongo.Database) (interfaces.UserRepository, error) {
	switch f.dbType {
	case PostgreSQL:
		if pgDB == nil {
			return nil, fmt.Errorf("PostgreSQL connection is required for PostgreSQL repository")
		}
		return postgresRepo.NewUserRepository(pgDB, f.logger), nil
	case MongoDB:
		if mongoDB == nil {
			return nil, fmt.Errorf("MongoDB connection is required for MongoDB repository")
		}
		return mongoRepo.NewUserRepository(mongoDB, f.logger), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", f.dbType)
	}
}

// CreateTodoRepository creates a todo repository based on database type
func (f *RepositoryFactory) CreateTodoRepository(pgDB *pgxpool.Pool, mongoDB *mongo.Database) (interfaces.TodoRepository, error) {
	switch f.dbType {
	case PostgreSQL:
		if pgDB == nil {
			return nil, fmt.Errorf("PostgreSQL connection is required for PostgreSQL repository")
		}
		return postgresRepo.NewTodoRepository(pgDB, f.logger), nil
	case MongoDB:
		if mongoDB == nil {
			return nil, fmt.Errorf("MongoDB connection is required for MongoDB repository")
		}
		return mongoRepo.NewTodoRepository(mongoDB, f.logger), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", f.dbType)
	}
}

// CreateRepositories creates all repositories based on database type
func (f *RepositoryFactory) CreateRepositories(pgDB *pgxpool.Pool, mongoDB *mongo.Database) (*interfaces.Repositories, error) {
	userRepo, err := f.CreateUserRepository(pgDB, mongoDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create user repository: %w", err)
	}

	todoRepo, err := f.CreateTodoRepository(pgDB, mongoDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create todo repository: %w", err)
	}

	return &interfaces.Repositories{
		User: userRepo,
		Todo: todoRepo,
	}, nil
}

// GetDatabaseType returns the current database type
func (f *RepositoryFactory) GetDatabaseType() DatabaseType {
	return f.dbType
}

// SetDatabaseType sets the database type
func (f *RepositoryFactory) SetDatabaseType(dbType DatabaseType) {
	f.dbType = dbType
}
