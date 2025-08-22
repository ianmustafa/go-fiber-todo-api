# Go Fiber Todo API

A comprehensive, production-ready Todo API built with Go Fiber framework, featuring clean architecture, JWT authentication, and multiple database support.

## 🚀 Features

- **Clean Architecture**: Layered design with clear separation of concerns
- **Multiple Database Support**: PostgreSQL and MongoDB with easy switching
- **JWT Authentication**: Secure authentication with access and refresh tokens
- **Session Management**: Redis-based session storage
- **ULID Support**: Using ULIDs for better performance and uniqueness
- **Soft Delete**: Non-destructive data deletion
- **Request Validation**: Comprehensive input validation
- **Rate Limiting**: Built-in rate limiting middleware
- **Health Checks**: Multiple health check endpoints
- **Swagger Documentation**: Auto-generated API documentation
- **Docker Support**: Complete containerization with Docker Compose
- **Unit Testing**: Comprehensive test suite with mocks
- **Structured Logging**: JSON-structured logging with Zerolog
- **Graceful Shutdown**: Proper server shutdown handling

## 📋 Table of Contents

- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Database Setup](#database-setup)
- [Running the Application](#running-the-application)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [Docker Deployment](#docker-deployment)
- [Project Structure](#project-structure)
- [Contributing](#contributing)

## 🏗️ Architecture

This project follows Clean Architecture principles with the following layers:

```
┌─────────────────┐
│   Handlers      │  ← HTTP handlers (controllers)
├─────────────────┤
│   Services      │  ← Business logic
├─────────────────┤
│  Repositories   │  ← Data access layer
├─────────────────┤
│   Database      │  ← PostgreSQL/MongoDB
└─────────────────┘
```

### Key Components

- **Handlers**: HTTP request/response handling
- **Services**: Business logic and orchestration
- **Repositories**: Data persistence abstraction
- **Models**: Data structures and validation
- **Middleware**: Cross-cutting concerns (auth, logging, etc.)
- **Config**: Configuration management

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (if not using Docker)
- MongoDB 6+ (if not using Docker)
- Redis 7+ (if not using Docker)
- Make (optional, for using Makefile commands)

## 🛠️ Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-fiber-todo-api
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Install development tools**
   ```bash
   # Install migration tool
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   
   # Install SQLC (for PostgreSQL)
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   
   # Install Swagger generator
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

## ⚙️ Configuration

The application uses environment variables for configuration. Create a `.env` file in the root directory:

```env
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=9000
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
SERVER_ENVIRONMENT=development

# Database Configuration
DATABASE_DRIVER=postgres  # or mongodb
DATABASE_POSTGRES_URL=postgres://user:password@localhost:5432/todoapp?sslmode=disable
DATABASE_MONGO_URL=mongodb://localhost:27017/todoapp
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5

# Redis Configuration
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-at-least-32-characters-long
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
JWT_ISSUER=go-fiber-todo-api

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## 🗄️ Database Setup

### PostgreSQL Setup

1. **Start PostgreSQL** (using Docker):
   ```bash
   docker run --name postgres-todo \
     -e POSTGRES_USER=user \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=todoapp \
     -p 5432:5432 \
     -d postgres:15
   ```

2. **Run migrations**:
   ```bash
   make migrate-up
   # or manually:
   migrate -path migrations/postgres -database "postgres://user:password@localhost:5432/todoapp?sslmode=disable" up
   ```

### MongoDB Setup

1. **Start MongoDB** (using Docker):
   ```bash
   docker run --name mongo-todo \
     -p 27017:27017 \
     -d mongo:6
   ```

### Redis Setup

1. **Start Redis** (using Docker):
   ```bash
   docker run --name redis-todo \
     -p 6379:6379 \
     -d redis:7-alpine
   ```

## 🚀 Running the Application

### Using Make Commands

```bash
# Run the application
make run

# Run with hot reload (requires air)
make dev

# Build the application
make build

# Run tests
make test

# Generate Swagger docs
make swagger

# Run linter
make lint
```

### Manual Commands

```bash
# Run the application
go run main.go

# Build the application
go build -o bin/app main.go

# Run tests
go test ./...

# Generate Swagger documentation
swag init
```

## 📚 API Documentation

The API documentation is automatically generated using Swagger and is available at:

- **Swagger UI**: `http://localhost:9000/swagger/index.html`
- **JSON**: `http://localhost:9000/swagger/doc.json`
- **YAML**: `http://localhost:9000/swagger/swagger.yaml`

### Main Endpoints

#### Authentication
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout user
- `GET /api/v1/auth/me` - Get current user profile

#### Todos
- `GET /api/v1/todos` - List todos with pagination
- `POST /api/v1/todos` - Create a new todo
- `GET /api/v1/todos/{id}` - Get todo by ID
- `PUT /api/v1/todos/{id}` - Update todo
- `DELETE /api/v1/todos/{id}` - Delete todo
- `PATCH /api/v1/todos/{id}/status` - Update todo status
- `GET /api/v1/todos/search` - Search todos
- `GET /api/v1/todos/overdue` - Get overdue todos
- `GET /api/v1/todos/stats` - Get todo statistics

#### Health Checks
- `GET /health` - General health check
- `GET /health/ready` - Readiness probe
- `GET /health/live` - Liveness probe

## 🧪 Testing

The project includes comprehensive unit tests with mocks.

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./internal/services -v

# Run tests with race detection
go test -race ./...
```

### Test Structure

- **Unit Tests**: Test individual components in isolation
- **Mocks**: Generated mocks for external dependencies
- **Test Utilities**: Helper functions for testing
- **Integration Tests**: End-to-end testing scenarios

## 🐳 Docker Deployment

### Using Docker Compose (Recommended)

1. **Start all services**:
   ```bash
   docker-compose up -d
   ```

2. **View logs**:
   ```bash
   docker-compose logs -f app
   ```

3. **Stop services**:
   ```bash
   docker-compose down
   ```

### Using Docker Only

1. **Build the image**:
   ```bash
   docker build -t go-fiber-todo-api .
   ```

2. **Run the container**:
   ```bash
   docker run -p 9000:9000 \
     -e DATABASE_POSTGRES_URL="postgres://user:password@host:5432/todoapp" \
     -e REDIS_URL="redis://host:6379/0" \
     -e JWT_SECRET="your-secret-key" \
     go-fiber-todo-api
   ```

## 📁 Project Structure

```
.
├── cmd/                   # Application entrypoints
├── docs/                  # Swagger documentation
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   ├── database/          # Database connections
│   │   ├── mongodb/       # MongoDB connection
│   │   └── postgres/      # PostgreSQL connection
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Data models
│   ├── repository/        # Data access layer
│   │   ├── interfaces/    # Repository interfaces
│   │   ├── mongodb/       # MongoDB implementations
│   │   └── postgres/      # PostgreSQL implementations
│   ├── services/          # Business logic
│   └── mocks/             # Test mocks
├── migrations/            # Database migrations
│   └── postgres/          # PostgreSQL migrations
├── scripts/               # Build and deployment scripts
├── testdata/              # Test data files
├── docker-compose.yml     # Docker Compose configuration
├── Dockerfile             # Docker image definition
├── Makefile               # Build automation
├── main.go                # Application entry point
└── README.md              # This file
```

## 🔧 Development

### Code Generation

```bash
# Generate SQLC code (PostgreSQL)
make sqlc-generate

# Generate Swagger documentation
make swagger

# Generate mocks
make mocks
```

### Database Migrations

```bash
# Create new migration
make migrate-create name=add_new_table

# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down

# Check migration status
make migrate-version
```

### Code Quality

```bash
# Run linter
make lint

# Format code
make fmt

# Run security check
make security
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Write comprehensive tests for new features
- Update documentation for API changes
- Use conventional commit messages
- Ensure all tests pass before submitting PR

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) - Express-inspired web framework
- [SQLC](https://sqlc.dev/) - Type-safe SQL code generation
- [Testify](https://github.com/stretchr/testify) - Testing toolkit
- [Zerolog](https://github.com/rs/zerolog) - Structured logging
- [Viper](https://github.com/spf13/viper) - Configuration management

## 📞 Support

If you have any questions or need help, please:

1. Check the [documentation](#api-documentation)
2. Search existing [issues](../../issues)
3. Create a new [issue](../../issues/new) if needed

---

**Happy coding! 🎉**
