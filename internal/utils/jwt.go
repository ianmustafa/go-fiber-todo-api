package utils

import (
	"fmt"
	"time"

	"go-fiber/internal/config"
	"go-fiber/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	SessionID string `json:"session_id"`
	Type      string `json:"type"`
	jwt.RegisteredClaims
}

// JWTService handles JWT operations
type JWTService struct {
	config config.JWTConfig
}

// NewJWTService creates a new JWT service
func NewJWTService(config config.JWTConfig) *JWTService {
	return &JWTService{
		config: config,
	}
}

// GenerateAccessToken generates an access token for the user
func (j *JWTService) GenerateAccessToken(user *models.User, sessionID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(j.config.AccessExpiry)

	claims := &JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		SessionID: sessionID,
		Type:      models.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

// GenerateRefreshToken generates a refresh token for the user
func (j *JWTService) GenerateRefreshToken(user *models.User, sessionID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(j.config.RefreshExpiry)

	claims := &JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		SessionID: sessionID,
		Type:      models.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (j *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != models.TokenTypeAccess {
		return nil, fmt.Errorf("invalid token type: expected access, got %s", claims.Type)
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (j *JWTService) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != models.TokenTypeRefresh {
		return nil, fmt.Errorf("invalid token type: expected refresh, got %s", claims.Type)
	}

	return claims, nil
}

// GenerateSessionID generates a new session ID
func GenerateSessionID() string {
	return ulid.Make().String()
}

// ExtractTokenFromHeader extracts the token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("authorization header must start with 'Bearer '")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}
