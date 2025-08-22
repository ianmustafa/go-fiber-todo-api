# Go Fiber Implementation Guide

## Dependencies and Go Modules

### Core Dependencies
```go
// go.mod
module go-fiber

go 1.21

require (
    // Web Framework
    github.com/gofiber/fiber/v2 v2.52.0
    github.com/gofiber/swagger v0.1.14
    
    // Configuration
    github.com/spf13/viper v1.18.2
    github.com/joho/godotenv v1.5.1
    
    // Database - PostgreSQL
    github.com/jackc/pgx/v5 v5.5.1
    github.com/pressly/goose/v3 v3.17.0
    
    // Database - MongoDB
    go.mongodb.org/mongo-driver v1.13.1
    
    // Redis
    github.com/redis/go-redis/v9 v9.4.0
    
    // Authentication
    github.com/golang-jwt/jwt/v5 v5.2.0
    golang.org/x/crypto v0.18.0
    
    // Validation
    github.com/go-playground/validator/v10 v10.16.0
    
    // Logging
    github.com/rs/zerolog v1.32.0
    
    // ULID
    github.com/oklog/ulid/v2 v2.1.0
    
    // Testing
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
    
    // Code Generation
    github.com/kyleconroy/sqlc v1.25.0
)
```

## Configuration Implementation

### Environment Configuration Structure
```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type ServerConfig struct {
    Host         string        `mapstructure:"host" default:"localhost"`
    Port         int           `mapstructure:"port" default:"9000"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout" default:"10s"`
    WriteTimeout time.Duration `mapstructure:"write_timeout" default:"10s"`
    Environment  string        `mapstructure:"environment" default:"development"`
}

type DatabaseConfig struct {
    Driver      string `mapstructure:"driver" default:"postgres"`
    PostgresURL string `mapstructure:"postgres_url"`
    MongoURL    string `mapstructure:"mongo_url"`
    MaxOpenConns int   `mapstructure:"max_open_conns" default:"25"`
    MaxIdleConns int   `mapstructure:"max_idle_conns" default:"5"`
}

type RedisConfig struct {
    URL      string `mapstructure:"url" default:"redis://localhost:6379/0"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db" default:"0"`
}

type JWTConfig struct {
    Secret          string        `mapstructure:"secret"`
    AccessExpiry    time.Duration `mapstructure:"access_expiry" default:"15m"`
    RefreshExpiry   time.Duration `mapstructure:"refresh_expiry" default:"168h"`
    Issuer          string        `mapstructure:"issuer" default:"go-fiber"`
}

type RateLimitConfig struct {
    Requests int           `mapstructure:"requests" default:"100"`
    Window   time.Duration `mapstructure:"window" default:"1m"`
}
```

## Database Schema Design

### PostgreSQL Schema (with ULID support and soft delete)
```sql
-- migrations/postgres/20250822222550_initial_schema.sql
-- +goose Up

-- Create ULID extension from ghcr.io/kavist/postgres-ulid
CREATE EXTENSION IF NOT EXISTS "ulid";

-- Users table with soft delete
CREATE TABLE users (
    id ulid PRIMARY KEY DEFAULT gen_ulid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    image VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Todos table with soft delete
CREATE TABLE todos (
    id ulid PRIMARY KEY DEFAULT gen_ulid(),
    user_id ulid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
    priority VARCHAR(10) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    due_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance (excluding soft deleted records)
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_todos_user_id ON todos(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_status ON todos(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_due_date ON todos(due_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_created_at ON todos(created_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_deleted_at ON todos(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS todos;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "ulid";
```

### SQLC Configuration
```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/repository/postgres/queries/"
    schema: "migrations/postgres/"
    gen:
      go:
        package: "queries"
        out: "internal/repository/postgres/queries"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_db_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
```

### SQLC Queries
```sql
-- internal/repository/postgres/queries/users.sql
-- name: CreateUser :one
INSERT INTO users (username, password_hash, email, image)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: UpdateUser :one
UPDATE users
SET username = $2, email = $3, image = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserImage :one
UPDATE users
SET image = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: GetUserTodos :many
SELECT * FROM todos 
WHERE user_id = $1 
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
```

## Repository Implementation Pattern

### Repository Interfaces (Separate Interface Folder)

#### User Repository Interface
```go
// internal/repository/interfaces/user.go
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

#### Todo Repository Interface
```go
// internal/repository/interfaces/todo.go
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

#### Repository Container
```go
// internal/repository/interfaces/repositories.go
type Repositories struct {
    User UserRepository
    Todo TodoRepository
}
```

#### Repository Factory Pattern
```go
// internal/repository/factory.go
type RepositoryFactory struct {
    pgConn    *pgxpool.Pool
    mongoConn *mongo.Database
    driver    string
    logger    zerolog.Logger
}

func NewRepositoryFactory(pgConn *pgxpool.Pool, mongoConn *mongo.Database, driver string, logger zerolog.Logger) *RepositoryFactory {
    return &RepositoryFactory{
        pgConn:    pgConn,
        mongoConn: mongoConn,
        driver:    driver,
        logger:    logger,
    }
}

func (f *RepositoryFactory) CreateUserRepository() interfaces.UserRepository {
    switch f.driver {
    case "mongodb":
        return mongodb.NewUserRepository(f.mongoConn, f.logger)
    default:
        return postgres.NewUserRepository(f.pgConn, f.logger)
    }
}

func (f *RepositoryFactory) CreateTodoRepository() interfaces.TodoRepository {
    switch f.driver {
    case "mongodb":
        return mongodb.NewTodoRepository(f.mongoConn, f.logger)
    default:
        return postgres.NewTodoRepository(f.pgConn, f.logger)
    }
}
```

### PostgreSQL Repository Implementation
```go
// internal/repository/postgres/user.go
type userRepository struct {
    db      *pgxpool.Pool
    queries *queries.Queries
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
    return &userRepository{
        db:      db,
        queries: queries.New(db),
    }
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
    var email, image *string
    if user.Email != "" {
        email = &user.Email
    }
    if user.Image != "" {
        image = &user.Image
    }
    
    dbUser, err := r.queries.CreateUser(ctx, queries.CreateUserParams{
        Username:     user.Username,
        PasswordHash: user.Password,
        Email:        email,
        Image:        image,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    result := &models.User{
        ID:        dbUser.ID,
        Username:  dbUser.Username,
        CreatedAt: dbUser.CreatedAt,
        UpdatedAt: dbUser.UpdatedAt,
    }
    
    if dbUser.Email != nil {
        result.Email = *dbUser.Email
    }
    if dbUser.Image != nil {
        result.Image = *dbUser.Image
    }
    
    return result, nil
}
```

## Authentication System Implementation

### JWT Service with Redis Session Management
```go
// internal/services/auth.go
type AuthService struct {
    userRepo     interfaces.UserRepository
    sessionStore SessionStore
    jwtConfig    *config.JWTConfig
    logger       zerolog.Logger
    bcryptCost   int
}

func NewAuthService(userRepo interfaces.UserRepository, sessionStore SessionStore, jwtConfig *config.JWTConfig, logger zerolog.Logger) *AuthService {
    return &AuthService{
        userRepo:     userRepo,
        sessionStore: sessionStore,
        jwtConfig:    jwtConfig,
        logger:       logger,
        bcryptCost:   bcrypt.DefaultCost,
    }
}

// Register handles user registration with validation
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
    // Check if username already exists
    exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
    if err != nil {
        return nil, fmt.Errorf("failed to check username existence: %w", err)
    }
    if exists {
        return nil, fmt.Errorf("username already exists")
    }

    // Check if email already exists (if provided)
    if req.Email != "" {
        exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
        if err != nil {
            return nil, fmt.Errorf("failed to check email existence: %w", err)
        }
        if exists {
            return nil, fmt.Errorf("email already exists")
        }
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    // Create user
    user := &models.User{
        Username: req.Username,
        Password: string(hashedPassword),
        Email:    req.Email,
    }

    createdUser, err := s.userRepo.Create(ctx, user)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    // Generate tokens and create session
    sessionID := s.generateSessionID()
    accessToken, err := s.generateAccessToken(createdUser.ID, createdUser.Username, sessionID)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }

    refreshToken, err := s.generateRefreshToken(createdUser.ID, createdUser.Username, sessionID)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }

    // Store session in Redis
    session := &models.Session{
        ID:        sessionID,
        UserID:    createdUser.ID,
        IsActive:  true,
        ExpiresAt: time.Now().Add(s.jwtConfig.RefreshExpiry),
    }

    if err := s.sessionStore.Set(ctx, sessionID, session, s.jwtConfig.RefreshExpiry); err != nil {
        return nil, fmt.Errorf("failed to create session: %w", err)
    }

    return &models.RegisterResponse{
        Message:      "User registered successfully",
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        User:         createdUser.ToResponse(),
    }, nil
}

// Login handles user authentication
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
    // Get user by username
    user, err := s.userRepo.GetByUsername(ctx, req.Username)
    if err != nil {
        s.logger.Error().Err(err).Str("username", req.Username).Msg("User not found")
        return nil, fmt.Errorf("invalid credentials")
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        s.logger.Error().Err(err).Str("username", req.Username).Msg("Invalid password")
        return nil, fmt.Errorf("invalid credentials")
    }

    // Generate session and tokens
    sessionID := s.generateSessionID()
    accessToken, err := s.generateAccessToken(user.ID, user.Username, sessionID)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }

    refreshToken, err := s.generateRefreshToken(user.ID, user.Username, sessionID)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }

    // Store session in Redis
    session := &models.Session{
        ID:        sessionID,
        UserID:    user.ID,
        IsActive:  true,
        ExpiresAt: time.Now().Add(s.jwtConfig.RefreshExpiry),
    }

    if err := s.sessionStore.Set(ctx, sessionID, session, s.jwtConfig.RefreshExpiry); err != nil {
        return nil, fmt.Errorf("failed to create session: %w", err)
    }

    return &models.LoginResponse{
        Message:      "Login successful",
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        User:         user.ToResponse(),
    }, nil
}
```

### Redis Session Management
```go
// internal/services/session.go
type RedisSessionStore struct {
    client redis.Cmdable
    logger zerolog.Logger
    prefix string
}

func NewRedisSessionStore(client redis.Cmdable, logger zerolog.Logger) *RedisSessionStore {
    return &RedisSessionStore{
        client: client,
        logger: logger,
        prefix: "session:",
    }
}

// Set stores a session in Redis
func (s *RedisSessionStore) Set(ctx context.Context, sessionID string, session *models.Session, expiration time.Duration) error {
    key := s.getKey(sessionID)

    // Serialize session to JSON
    data, err := json.Marshal(session)
    if err != nil {
        s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to marshal session")
        return fmt.Errorf("failed to marshal session: %w", err)
    }

    // Store in Redis with expiration
    if err := s.client.Set(ctx, key, data, expiration).Err(); err != nil {
        s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to store session in Redis")
        return fmt.Errorf("failed to store session: %w", err)
    }

    s.logger.Debug().Str("session_id", sessionID).Dur("expiration", expiration).Msg("Session stored successfully")
    return nil
}

// Get retrieves a session from Redis
func (s *RedisSessionStore) Get(ctx context.Context, sessionID string) (*models.Session, error) {
    key := s.getKey(sessionID)

    // Get from Redis
    data, err := s.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, fmt.Errorf("session not found")
        }
        s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get session from Redis")
        return nil, fmt.Errorf("failed to get session: %w", err)
    }

    // Deserialize session from JSON
    var session models.Session
    if err := json.Unmarshal([]byte(data), &session); err != nil {
        s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to unmarshal session")
        return nil, fmt.Errorf("failed to unmarshal session: %w", err)
    }

    s.logger.Debug().Str("session_id", sessionID).Msg("Session retrieved successfully")
    return &session, nil
}

// Delete removes a session from Redis
func (s *RedisSessionStore) Delete(ctx context.Context, sessionID string) error {
    key := s.getKey(sessionID)

    // Delete from Redis
    result, err := s.client.Del(ctx, key).Result()
    if err != nil {
        s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to delete session from Redis")
        return fmt.Errorf("failed to delete session: %w", err)
    }

    if result == 0 {
        s.logger.Warn().Str("session_id", sessionID).Msg("Session not found for deletion")
        return fmt.Errorf("session not found")
    }

    s.logger.Debug().Str("session_id", sessionID).Msg("Session deleted successfully")
    return nil
}

func (s *RedisSessionStore) getKey(sessionID string) string {
    return s.prefix + sessionID
}
```

## Middleware Implementation

### Authentication Middleware
```go
// internal/middleware/auth.go
func AuthMiddleware(sessionSvc *services.SessionService, jwtConfig config.JWTConfig) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Authorization header required",
            })
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Bearer token required",
            })
        }
        
        // Parse and validate JWT
        token, err := jwt.ParseWithClaims(tokenString, &utils.Claims{}, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtConfig.Secret), nil
        })
        
        if err != nil || !token.Valid {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Invalid token",
            })
        }
        
        claims, ok := token.Claims.(*utils.Claims)
        if !ok {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Invalid token claims",
            })
        }
        
        // Validate session in Redis
        if err := sessionSvc.ValidateSession(c.Context(), claims.UserID, claims.SessionID); err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Session expired or invalid",
            })
        }
        
        // Set user context
        c.Locals("userID", claims.UserID)
        c.Locals("sessionID", claims.SessionID)
        
        return c.Next()
    }
}
```

### Request Validation Middleware
```go
// internal/middleware/validation.go
func ValidateRequest[T any]() fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req T
        
        if err := c.BodyParser(&req); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "Invalid request body",
                "details": err.Error(),
            })
        }
        
        validate := validator.New()
        if err := validate.Struct(&req); err != nil {
            var validationErrors []string
            for _, err := range err.(validator.ValidationErrors) {
                validationErrors = append(validationErrors, fmt.Sprintf(
                    "Field '%s' failed validation: %s",
                    err.Field(),
                    err.Tag(),
                ))
            }
            
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "Validation failed",
                "details": validationErrors,
            })
        }
        
        c.Locals("validatedRequest", req)
        return c.Next()
    }
}
```

## API Handler Implementation

### Todo Handlers
```go
// internal/handlers/todo.go
type TodoHandler struct {
    todoSvc *services.TodoService
    logger  zerolog.Logger
}

// @Summary Create a new todo
// @Description Create a new todo item for the authenticated user
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body models.CreateTodoRequest true "Todo data"
// @Success 201 {object} models.Todo
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/todos [post]
func (h *TodoHandler) CreateTodo(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    req := c.Locals("validatedRequest").(models.CreateTodoRequest)
    
    todo, err := h.todoSvc.CreateTodo(c.Context(), userID, &req)
    if err != nil {
        h.logger.Error().Err(err).Msg("Failed to create todo")
        return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
            Error: "Failed to create todo",
        })
    }
    
    return c.Status(fiber.StatusCreated).JSON(todo)
}

// @Summary Get user todos
// @Description Get paginated list of todos for the authenticated user
// @Tags todos
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param status query string false "Filter by status"
// @Success 200 {object} models.TodoListResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/todos [get]
func (h *TodoHandler) GetTodos(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    
    limit := c.QueryInt("limit", 10)
    offset := c.QueryInt("offset", 0)
    status := c.Query("status")
    
    todos, total, err := h.todoSvc.GetUserTodos(c.Context(), userID, limit, offset, status)
    if err != nil {
        h.logger.Error().Err(err).Msg("Failed to get todos")
        return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
            Error: "Failed to get todos",
        })
    }
    
    return c.JSON(models.TodoListResponse{
        Todos: todos,
        Total: total,
        Limit: limit,
        Offset: offset,
    })
}
```

## Testing Strategy Implementation

### Service Layer Testing
```go
// internal/services/todo_test.go
func TestTodoService_CreateTodo(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockRepo := mocks.NewMockTodoRepository(ctrl)
    logger := zerolog.New(os.Stdout)
    
    service := services.NewTodoService(mockRepo, logger)
    
    tests := []struct {
        name    string
        userID  string
        req     *models.CreateTodoRequest
        setup   func()
        wantErr bool
    }{
        {
            name:   "successful creation",
            userID: "user123",
            req: &models.CreateTodoRequest{
                Title:       "Test Todo",
                Description: "Test Description",
                Priority:    "high",
            },
            setup: func() {
                mockRepo.EXPECT().
                    Create(gomock.Any(), gomock.Any()).
                    Return(&models.Todo{
                        ID:          "todo123",
                        UserID:      "user123",
                        Title:       "Test Todo",
                        Description: "Test Description",
                        Status:      "pending",
                        Priority:    "high",
                    }, nil)
            },
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            
            todo, err := service.CreateTodo(context.Background(), tt.userID, tt.req)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, todo)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, todo)
                assert.Equal(t, tt.req.Title, todo.Title)
            }
        })
    }
}
```

## Docker Configuration

### Multi-stage Dockerfile
```dockerfile
# docker/Dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/go-fiber

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

# Expose port
EXPOSE 9000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9000/health || exit 1

CMD ["./main"]
```

### Docker Compose Configuration
```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: docker/Dockerfile
    ports:
      - "9000:9000"
    environment:
      - DB_DRIVER=postgres
      - POSTGRES_URL=postgresql://postgres:password@postgres:5432/go_fiber?sslmode=disable
      - REDIS_URL=redis://redis:6379/0
      - JWT_SECRET=your-super-secret-jwt-key
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - app-network

  postgres:
    image: ghcr.io/kavist/postgres-ulid:latest
    environment:
      POSTGRES_DB: go_fiber
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network

  mongodb:
    image: mongo:7
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: password
      MONGO_INITDB_DATABASE: go_fiber
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    healthcheck:
      test: echo 'db.runCommand("ping").ok' | mongosh localhost:27017/test --quiet
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network

volumes:
  postgres_data:
  mongodb_data:
  redis_data:

networks:
  app-network:
    driver: bridge
```

This implementation guide provides detailed technical specifications for building the production-ready Go Fiber application with clean architecture, multiple database support, and comprehensive authentication system.