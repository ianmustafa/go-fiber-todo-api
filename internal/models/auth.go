package models

import (
	"time"
)

// LoginRequest represents the request to login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginByEmailRequest represents the request to login by email
type LoginByEmailRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	ExpiresAt    time.Time     `json:"expiresAt"`
	User         *UserResponse `json:"user"`
}

// RefreshTokenRequest represents the request to refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// RefreshTokenResponse represents the response after token refresh
type RefreshTokenResponse struct {
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// RegisterRequest represents the request to register a new user
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=100"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Image    string `json:"image,omitempty" validate:"omitempty,url"`
}

// RegisterResponse represents the response after successful registration
type RegisterResponse struct {
	User    *UserResponse `json:"user"`
	Message string        `json:"message"`
}

// LogoutRequest represents the request to logout
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken,omitempty"`
}

// LogoutResponse represents the response after logout
type LogoutResponse struct {
	Message string `json:"message"`
}

// AuthUserResponse represents the authenticated user response
type AuthUserResponse struct {
	User *UserResponse `json:"user"`
}

// Claims represents JWT claims
type Claims struct {
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	SessionID string `json:"sessionId"`
	Type      string `json:"type"` // "access" or "refresh"
}

// TokenType constants
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	IsActive  bool      `json:"isActive"`
}
