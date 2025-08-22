package models

import "time"

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Bad Request"`
	Message string `json:"message" example:"Invalid input data."`
	Details string `json:"details,omitempty" example:"Validation failed."`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully."`
}

// TodoResponse represents a todo response
type TodoResponse struct {
	Message string `json:"message" example:"Todo retrieved successfully."`
	Data    *Todo  `json:"data"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Message      string `json:"message" example:"Login successful."`
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User         *User  `json:"user"`
}

// TokenResponse represents a token refresh response
type TokenResponse struct {
	Message     string `json:"message" example:"Token refreshed successfully."`
	AccessToken string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string                 `json:"status" example:"healthy"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version" example:"1.0.0"`
	Services  map[string]ServiceInfo `json:"services"`
}

// ServiceInfo represents the status of a service
type ServiceInfo struct {
	Status       string `json:"status" example:"healthy"`
	ResponseTime string `json:"responseTime" example:"5ms"`
	Error        string `json:"error,omitempty" example:"Connection failed."`
}
