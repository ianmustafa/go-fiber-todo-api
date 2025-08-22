# API Request Examples

This document provides example API requests for the Go Fiber Todo API.

## Authentication

### Register a New User

```bash
curl -X POST http://localhost:9000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "securepassword123",
    "email": "john@example.com"
  }'
```

**Response:**
```json
{
  "message": "User registered successfully",
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "01HKQM7X8YZABCDEFGHIJKLMNO",
    "username": "johndoe",
    "email": "john@example.com",
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:30:00Z"
  }
}
```

### Login User

```bash
curl -X POST http://localhost:9000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "securepassword123"
  }'
```

**Response:**
```json
{
  "message": "Login successful",
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "01HKQM7X8YZABCDEFGHIJKLMNO",
    "username": "johndoe",
    "email": "john@example.com",
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:30:00Z"
  }
}
```

### Refresh Access Token

```bash
curl -X POST http://localhost:9000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Response:**
```json
{
  "message": "Token refreshed successfully",
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Logout User

```bash
curl -X POST http://localhost:9000/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "message": "Logout successful"
}
```

### Get Current User Profile

```bash
curl -X GET http://localhost:9000/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "message": "User profile retrieved successfully",
  "user": {
    "id": "01HKQM7X8YZABCDEFGHIJKLMNO",
    "username": "johndoe",
    "email": "john@example.com",
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:30:00Z"
  }
}
```

## Todo Management

### Create a New Todo

```bash
curl -X POST http://localhost:9000/api/v1/todos \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete project documentation",
    "description": "Write comprehensive API documentation and examples",
    "priority": "high",
    "dueDate": "2024-01-20T18:00:00Z"
  }'
```

**Response:**
```json
{
  "id": "01HKQM8Y9ZABCDEFGHIJKLMNOP",
  "userId": "01HKQM7X8YZABCDEFGHIJKLMNO",
  "title": "Complete project documentation",
  "description": "Write comprehensive API documentation and examples",
  "status": "pending",
  "priority": "high",
  "dueDate": "2024-01-20T18:00:00Z",
  "createdAt": "2024-01-15T10:35:00Z",
  "updatedAt": "2024-01-15T10:35:00Z"
}
```

### Get All Todos (with pagination)

```bash
curl -X GET "http://localhost:9000/api/v1/todos?limit=10&offset=0" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "todos": [
    {
      "id": "01HKQM8Y9ZABCDEFGHIJKLMNOP",
      "userId": "01HKQM7X8YZABCDEFGHIJKLMNO",
      "title": "Complete project documentation",
      "description": "Write comprehensive API documentation and examples",
      "status": "pending",
      "priority": "high",
      "dueDate": "2024-01-20T18:00:00Z",
      "createdAt": "2024-01-15T10:35:00Z",
      "updatedAt": "2024-01-15T10:35:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

### Get Todo by ID

```bash
curl -X GET http://localhost:9000/api/v1/todos/01HKQM8Y9ZABCDEFGHIJKLMNOP \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "id": "01HKQM8Y9ZABCDEFGHIJKLMNOP",
  "userId": "01HKQM7X8YZABCDEFGHIJKLMNO",
  "title": "Complete project documentation",
  "description": "Write comprehensive API documentation and examples",
  "status": "pending",
  "priority": "high",
  "dueDate": "2024-01-20T18:00:00Z",
  "createdAt": "2024-01-15T10:35:00Z",
  "updatedAt": "2024-01-15T10:35:00Z"
}
```

### Update Todo

```bash
curl -X PUT http://localhost:9000/api/v1/todos/01HKQM8Y9ZABCDEFGHIJKLMNOP \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete project documentation (Updated)",
    "description": "Write comprehensive API documentation, examples, and README",
    "status": "in_progress",
    "priority": "high"
  }'
```

**Response:**
```json
{
  "id": "01HKQM8Y9ZABCDEFGHIJKLMNOP",
  "userId": "01HKQM7X8YZABCDEFGHIJKLMNO",
  "title": "Complete project documentation (Updated)",
  "description": "Write comprehensive API documentation, examples, and README",
  "status": "in_progress",
  "priority": "high",
  "dueDate": "2024-01-20T18:00:00Z",
  "createdAt": "2024-01-15T10:35:00Z",
  "updatedAt": "2024-01-15T10:40:00Z"
}
```

### Update Todo Status

```bash
curl -X PATCH http://localhost:9000/api/v1/todos/01HKQM8Y9ZABCDEFGHIJKLMNOP/status \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed"
  }'
```

**Response:**
```json
{
  "message": "Todo status updated successfully",
  "status": "completed"
}
```

### Search Todos

```bash
curl -X GET "http://localhost:9000/api/v1/todos/search?q=documentation&limit=10&offset=0" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "todos": [
    {
      "id": "01HKQM8Y9ZABCDEFGHIJKLMNOP",
      "userId": "01HKQM7X8YZABCDEFGHIJKLMNO",
      "title": "Complete project documentation (Updated)",
      "description": "Write comprehensive API documentation, examples, and README",
      "status": "completed",
      "priority": "high",
      "dueDate": "2024-01-20T18:00:00Z",
      "createdAt": "2024-01-15T10:35:00Z",
      "updatedAt": "2024-01-15T10:45:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

### Get Overdue Todos

```bash
curl -X GET "http://localhost:9000/api/v1/todos/overdue?limit=10&offset=0" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Get Todo Statistics

```bash
curl -X GET http://localhost:9000/api/v1/todos/stats \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "stats": {
    "pending": 5,
    "in_progress": 3,
    "completed": 12
  }
}
```

### Delete Todo

```bash
curl -X DELETE http://localhost:9000/api/v1/todos/01HKQM8Y9ZABCDEFGHIJKLMNOP \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:** `204 No Content`

## Health Checks

### General Health Check

```bash
curl -X GET http://localhost:9000/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:50:00Z",
  "version": "1.0.0",
  "services": {
    "postgresql": {
      "status": "healthy",
      "responseTime": "2ms"
    },
    "redis": {
      "status": "healthy",
      "responseTime": "1ms"
    }
  }
}
```

### Readiness Check

```bash
curl -X GET http://localhost:9000/health/ready
```

### Liveness Check

```bash
curl -X GET http://localhost:9000/health/live
```

**Response:**
```json
{
  "status": "alive",
  "timestamp": "2024-01-15T10:50:00Z",
  "version": "1.0.0"
}
```

## Error Responses

### 400 Bad Request

```json
{
  "error": "Bad Request",
  "message": "Invalid request body",
  "details": "JSON parsing error"
}
```

### 401 Unauthorized

```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired token"
}
```

### 404 Not Found

```json
{
  "error": "Not Found",
  "message": "Todo not found"
}
```

### 409 Conflict

```json
{
  "error": "Conflict",
  "message": "Username already exists"
}
```

### 422 Validation Error

```json
{
  "error": "Validation Error",
  "message": "Invalid input data",
  "details": "Title is required and must be between 1 and 200 characters"
}
```

### 429 Too Many Requests

```json
{
  "error": "Too Many Requests",
  "message": "Rate limit exceeded"
}
```

### 500 Internal Server Error

```json
{
  "error": "Internal Server Error",
  "message": "An unexpected error occurred"
}
```

## Notes

- All timestamps are in ISO 8601 format (UTC)
- ULIDs are used for all entity IDs
- JWT tokens expire after 15 minutes (access) and 7 days (refresh)
- Rate limiting is applied per IP address
- All API endpoints use JSON for request/response bodies
- Authentication is required for all todo endpoints
- Soft delete is used for todos (they are not permanently deleted)