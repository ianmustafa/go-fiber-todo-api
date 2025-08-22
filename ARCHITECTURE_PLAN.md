# Go Fiber Production-Ready Architecture Plan

## Project Overview
A production-ready Go application using Fiber framework with clean architecture, supporting multiple databases (PostgreSQL, MongoDB), JWT authentication with Redis sessions, and comprehensive Todo/Task management system.

## Technology Stack
- **Go Version**: 1.21+
- **Web Framework**: Fiber v2
- **Databases**: PostgreSQL (with ULID support), MongoDB
- **Cache/Sessions**: Redis
- **Authentication**: JWT with Redis session tracking
- **Configuration**: Viper + godotenv
- **Logging**: zerolog (structured logging)
- **Validation**: go-playground/validator
- **Database Tools**: 
  - PostgreSQL: pgx (driver), sqlc (code generation), goose (migrations)
  - MongoDB: official mongo-go-driver
- **Testing**: testing + testify + mocks
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker + docker-compose

## Project Structure

```
go-fiber/
├── cmd/
│   └── go-fiber/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go              # Configuration struct and loading
│   │   └── test_config.go         # Test configuration utilities
│   ├── database/
│   │   ├── postgres/
│   │   │   └── connection.go      # PostgreSQL connection setup
│   │   ├── mongodb/
│   │   │   └── connection.go      # MongoDB connection setup
│   │   └── redis/                 # Redis connection utilities
│   ├── server/
│   │   └── server.go              # Fiber server setup and configuration
│   ├── handlers/
│   │   ├── auth.go                # Authentication handlers
│   │   ├── todo.go                # Todo CRUD handlers
│   │   ├── health.go              # Health check handler
│   │   └── todo_test.go           # Handler unit tests
│   ├── middleware/
│   │   ├── auth.go                # JWT + Redis session middleware
│   │   ├── logging.go             # Request logging middleware
│   │   ├── cors.go                # CORS middleware
│   │   └── ratelimit.go           # Rate limiting middleware
│   ├── services/
│   │   ├── auth.go                # Authentication business logic
│   │   ├── session.go             # Session management logic
│   │   └── auth_test.go           # Service unit tests
│   ├── repository/
│   │   ├── interfaces/
│   │   │   ├── user.go            # User repository interface
│   │   │   ├── todo.go            # Todo repository interface
│   │   │   └── repositories.go    # Repository container
│   │   ├── postgres/
│   │   │   ├── user.go            # PostgreSQL user repository
│   │   │   ├── todo.go            # PostgreSQL todo repository
│   │   │   └── queries/           # sqlc generated queries
│   │   ├── mongodb/
│   │   │   ├── user.go            # MongoDB user repository
│   │   │   └── todo.go            # MongoDB todo repository
│   │   ├── factory.go             # Repository factory pattern
│   │   └── mocks/                 # Mock implementations for testing
│   ├── models/
│   │   ├── user.go                # User domain models
│   │   ├── todo.go                # Todo domain models
│   │   ├── auth.go                # Authentication models
│   │   └── response.go            # API response models
│   ├── mocks/                     # Test mocks
│   │   ├── user_repository.go     # User repository mock
│   │   ├── todo_repository.go     # Todo repository mock
│   │   └── session_store.go       # Session store mock
│   └── utils/
│       ├── jwt.go                 # JWT utilities
│       ├── password.go            # Password hashing utilities
│       └── response.go            # Standard API response utilities
├── migrations/
│   ├── postgres/                  # PostgreSQL migrations with ULID support
│   └── mongodb/                   # MongoDB migrations (if needed)
├── docs/                          # Generated Swagger documentation
├── examples/                      # API examples and collections
│   ├── api-requests.md           # cURL examples
│   └── postman-collection.json   # Postman collection
├── scripts/                       # Build and deployment scripts
├── docker/
│   └── Dockerfile                 # Application Dockerfile
├── docker-compose.yml             # Development environment
├── .env.example                   # Environment variables template
├── .gitignore                     # Git ignore rules
├── Makefile                       # Build automation
├── sqlc.yaml                      # SQLC configuration
├── main.go                        # Application entry point (root level)
├── go.mod                         # Go modules
├── go.sum                         # Go modules checksum
└── README.md                      # Project documentation
```

## Architecture Patterns

### Clean Architecture Layers
1. **Presentation Layer** (`/handlers`): HTTP handlers, request/response handling
2. **Application Layer** (`/services`): Business logic, use cases
3. **Domain Layer** (`/models`): Core business entities and rules
4. **Infrastructure Layer** (`/repository`, `/database`): Data persistence, external services

### Repository Pattern
- Interface-based design for database abstraction
- Pluggable database drivers (PostgreSQL/MongoDB)
- Easy testing with mock implementations

### Dependency Injection
- Constructor-based dependency injection
- Interface segregation for better testability
- Configuration-driven database selection

## API Design

### Authentication Endpoints
```
POST /api/v1/auth/login     # User login (returns JWT)
POST /api/v1/auth/logout    # User logout (invalidates session)
GET  /api/v1/auth/user      # Get current user info
POST /api/v1/auth/refresh   # Refresh JWT token
```

### Todo Management Endpoints
```
GET    /api/v1/todos        # List user's todos (with pagination)
POST   /api/v1/todos        # Create new todo
GET    /api/v1/todos/:id    # Get specific todo
PUT    /api/v1/todos/:id    # Update todo
DELETE /api/v1/todos/:id    # Delete todo
PATCH  /api/v1/todos/:id/status # Update todo status
```

### User Management Endpoints
```
POST /api/v1/users          # Register new user
GET  /api/v1/users/profile  # Get user profile
PUT  /api/v1/users/profile  # Update user profile
```

### System Endpoints
```
GET /health                 # Health check
GET /api/v1/hello          # Simple hello endpoint
```

## Data Models

### User Model (with camelCase JSON and soft delete)
```go
type User struct {
    ID        string     `json:"id" db:"id"`
    Username  string     `json:"username" db:"username" validate:"required,min=3,max=50"`
    Password  string     `json:"-" db:"password_hash"`
    Email     string     `json:"email,omitempty" db:"email" validate:"omitempty,email"`
    Image     string     `json:"image,omitempty" db:"image" validate:"omitempty,url"`
    CreatedAt time.Time  `json:"createdAt" db:"created_at"`
    UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
    DeletedAt *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}
```

### Todo Model (with camelCase JSON and soft delete)
```go
type Todo struct {
    ID          string     `json:"id" db:"id"`
    UserID      string     `json:"userId" db:"user_id"`
    Title       string     `json:"title" db:"title" validate:"required,min=1,max=200"`
    Description string     `json:"description" db:"description"`
    Status      string     `json:"status" db:"status" validate:"required,oneof=pending in_progress completed"`
    Priority    string     `json:"priority" db:"priority" validate:"oneof=low medium high"`
    DueDate     *time.Time `json:"dueDate,omitempty" db:"due_date"`
    CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
    UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
    DeletedAt   *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}
```

## Authentication & Session Management

### JWT Strategy
- Stateless JWT tokens for API authentication
- Short-lived access tokens (15 minutes)
- Longer-lived refresh tokens (7 days)
- Redis-based session tracking for active sessions

### Session Management Flow
```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Redis
    participant DB

    Client->>API: POST /auth/login
    API->>DB: Validate credentials
    DB-->>API: User data
    API->>Redis: Store session (user_id:session_id)
    API-->>Client: JWT + session info

    Client->>API: Authenticated request
    API->>API: Validate JWT
    API->>Redis: Check session exists
    Redis-->>API: Session valid
    API-->>Client: Response

    Client->>API: POST /auth/logout
    API->>Redis: Delete session
    API-->>Client: Logout success
```

## Database Strategy

### PostgreSQL Schema (Primary)
- **ULID Support**: Using `ghcr.io/kavist/postgres-ulid` with `ulid` data type
- **ULID Generation**: `gen_ulid()` function for automatic ULID generation
- **Soft Delete**: `deleted_at` timestamp fields for non-destructive deletion
- **Proper Indexing**: Query optimization with strategic indexes
- **Foreign Key Constraints**: Data integrity enforcement
- **Migration-based Schema**: Goose migration system

### MongoDB Schema (Alternative)
- Document-based storage for flexible data models
- Embedded documents for related data
- Proper indexing for query performance
- Schema validation at application level

### Repository Interfaces (Separate Interface Folder)

#### User Repository Interface (`internal/repository/interfaces/user.go`)
```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) (*models.User, error)
    GetByID(ctx context.Context, id string) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    GetByUsername(ctx context.Context, username string) (*models.User, error)
    Update(ctx context.Context, user *models.User) (*models.User, error)
    Delete(ctx context.Context, id string) error
    UpdateImage(ctx context.Context, id, imageURL string) error
    UpdatePassword(ctx context.Context, id, hashedPassword string) error
    List(ctx context.Context, limit, offset int) ([]*models.User, int64, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    ExistsByUsername(ctx context.Context, username string) (bool, error)
}
```

#### Todo Repository Interface (`internal/repository/interfaces/todo.go`)
```go
type TodoRepository interface {
    Create(ctx context.Context, todo *models.Todo) (*models.Todo, error)
    GetByID(ctx context.Context, id string) (*models.Todo, error)
    GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Todo, int64, error)
    Update(ctx context.Context, todo *models.Todo) (*models.Todo, error)
    Delete(ctx context.Context, id string) error
    UpdateStatus(ctx context.Context, id, status string) error
    GetByStatus(ctx context.Context, userID, status string, limit, offset int) ([]*models.Todo, int64, error)
    GetByPriority(ctx context.Context, userID, priority string, limit, offset int) ([]*models.Todo, int64, error)
    GetOverdue(ctx context.Context, userID string, limit, offset int) ([]*models.Todo, int64, error)
    GetUpcoming(ctx context.Context, userID string, days int, limit, offset int) ([]*models.Todo, int64, error)
    Search(ctx context.Context, userID, query string, limit, offset int) ([]*models.Todo, int64, error)
    CountByStatus(ctx context.Context, userID string) (map[string]int64, error)
    MarkCompleted(ctx context.Context, id string) error
    BulkUpdateStatus(ctx context.Context, ids []string, status string) error
    DeleteCompleted(ctx context.Context, userID string) error
}
```

#### Repository Container (`internal/repository/interfaces/repositories.go`)
```go
type Repositories struct {
    User UserRepository
    Todo TodoRepository
}
```

#### Repository Factory Pattern (`internal/repository/factory.go`)
```go
type RepositoryFactory struct {
    pgConn    *pgxpool.Pool
    mongoConn *mongo.Database
    driver    string
    logger    zerolog.Logger
}

func (f *RepositoryFactory) CreateUserRepository() interfaces.UserRepository
func (f *RepositoryFactory) CreateTodoRepository() interfaces.TodoRepository
```

## Configuration Management

### Environment Variables
```bash
# Server Configuration
SERVER_PORT=9000
SERVER_HOST=localhost
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s

# Database Configuration
DB_DRIVER=postgres  # postgres or mongodb
POSTGRES_URL=postgresql://user:pass@localhost:5432/dbname
MONGODB_URL=mongodb://localhost:27017/dbname

# Redis Configuration
REDIS_URL=redis://localhost:6379/0

# JWT Configuration
JWT_SECRET=your-secret-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

## Testing Strategy

### Unit Tests
- Service layer testing with mocked repositories
- Repository testing with test databases
- Utility function testing
- JWT and authentication logic testing

### Integration Tests
- API endpoint testing
- Database integration testing
- Redis session testing
- End-to-end authentication flow testing

### Test Structure
```
internal/
├── services/
│   ├── auth_test.go
│   └── todo_test.go
├── repository/
│   ├── postgres/
│   │   ├── user_test.go
│   │   └── todo_test.go
│   └── mongodb/
│       ├── user_test.go
│       └── todo_test.go
└── handlers/
    ├── auth_test.go
    └── todo_test.go
```

## Deployment Strategy

### Docker Configuration
- Multi-stage Dockerfile for optimized production builds
- docker-compose for local development environment
- Health checks for all services
- Volume mounts for development hot-reload

### Production Considerations
- Graceful shutdown handling
- Proper error handling and logging
- Rate limiting and security headers
- Database connection pooling
- Redis connection management
- Environment-based configuration

## Security Measures

### Authentication Security
- Password hashing with bcrypt
- JWT token validation and expiration
- Session invalidation on logout
- Rate limiting on authentication endpoints

### API Security
- CORS configuration
- Request validation and sanitization
- SQL injection prevention (parameterized queries)
- NoSQL injection prevention
- Proper error handling (no sensitive data exposure)

## Monitoring & Observability

### Logging Strategy
- Structured logging with zerolog
- Request/response logging middleware
- Error logging with stack traces
- Performance metrics logging

### Health Checks
- Database connectivity checks
- Redis connectivity checks
- Service dependency health
- Custom health indicators

This architectural plan provides a comprehensive foundation for building a production-ready Go application with clean architecture principles, multiple database support, and robust authentication mechanisms.