package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"go-fiber/internal/config"
	"go-fiber/internal/models"
	"go-fiber/internal/repository/interfaces"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo     interfaces.UserRepository
	sessionStore SessionStore
	config       *config.JWTConfig
	logger       zerolog.Logger
	bcryptCost   int
}

// SessionStore interface for session management
type SessionStore interface {
	Set(ctx context.Context, sessionID string, session *models.Session, expiration time.Duration) error
	Get(ctx context.Context, sessionID string) (*models.Session, error)
	Delete(ctx context.Context, sessionID string) error
	DeleteUserSessions(ctx context.Context, userID string) error
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo interfaces.UserRepository,
	sessionStore SessionStore,
	config *config.JWTConfig,
	logger zerolog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		sessionStore: sessionStore,
		config:       config,
		logger:       logger,
		bcryptCost:   bcrypt.DefaultCost,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
	// Check if username already exists
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error().Err(err).Str("username", req.Username).Msg("Failed to check username existence.")
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists (if provided)
	if req.Email != "" {
		exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			s.logger.Error().Err(err).Str("email", req.Email).Msg("Failed to check email existence.")
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("email already exists")
		}
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash password.")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Image:    req.Image,
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Error().Err(err).Str("username", req.Username).Msg("Failed to create user.")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info().Str("user_id", createdUser.ID).Str("username", createdUser.Username).Msg("User registered successfully.")

	return &models.RegisterResponse{
		User:    createdUser.ToResponse(),
		Message: "User registered successfully",
	}, nil
}

// Login authenticates a user and returns JWT tokens
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error().Err(err).Str("username", req.Username).Msg("Failed to get user by username.")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := s.verifyPassword(user.Password, req.Password); err != nil {
		s.logger.Warn().Str("username", req.Username).Msg("Invalid password attempt.")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate session ID
	entropy := ulid.Monotonic(rand.Reader, 0)
	sessionID := ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()

	// Create session
	session := &models.Session{
		ID:        sessionID,
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.config.RefreshExpiry),
		IsActive:  true,
	}

	// Store session
	if err := s.sessionStore.Set(ctx, sessionID, session, s.config.RefreshExpiry); err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to store session.")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user.ID, user.Username, sessionID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", user.ID).Msg("Failed to generate access token.")
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID, user.Username, sessionID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", user.ID).Msg("Failed to generate refresh token.")
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	s.logger.Info().Str("user_id", user.ID).Str("username", user.Username).Msg("User logged in successfully.")

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.config.AccessExpiry),
		User:         user.ToResponse(),
	}, nil
}

// LoginByEmail authenticates a user by email and returns JWT tokens
func (s *AuthService) LoginByEmail(ctx context.Context, req *models.LoginByEmailRequest) (*models.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error().Err(err).Str("email", req.Email).Msg("Failed to get user by email.")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := s.verifyPassword(user.Password, req.Password); err != nil {
		s.logger.Warn().Str("email", req.Email).Msg("Invalid password attempt.")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate session ID
	entropy := ulid.Monotonic(rand.Reader, 0)
	sessionID := ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()

	// Create session
	session := &models.Session{
		ID:        sessionID,
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.config.RefreshExpiry),
		IsActive:  true,
	}

	// Store session
	if err := s.sessionStore.Set(ctx, sessionID, session, s.config.RefreshExpiry); err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to store session.")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user.ID, user.Username, sessionID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", user.ID).Msg("Failed to generate access token.")
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID, user.Username, sessionID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", user.ID).Msg("Failed to generate refresh token.")
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	s.logger.Info().Str("user_id", user.ID).Str("email", req.Email).Msg("User logged in successfully.")

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.config.AccessExpiry),
		User:         user.ToResponse(),
	}, nil
}

// RefreshToken generates new access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.RefreshTokenResponse, error) {
	// Parse and validate refresh token
	claims, err := s.validateToken(req.RefreshToken, models.TokenTypeRefresh)
	if err != nil {
		s.logger.Error().Err(err).Msg("Invalid refresh token.")
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get session
	session, err := s.sessionStore.Get(ctx, claims.SessionID)
	if err != nil {
		s.logger.Error().Err(err).Str("session_id", claims.SessionID).Msg("Failed to get session.")
		return nil, fmt.Errorf("invalid session")
	}

	// Check if session is active and not expired
	if !session.IsActive || time.Now().After(session.ExpiresAt) {
		s.logger.Warn().Str("session_id", claims.SessionID).Msg("Session is inactive or expired.")
		return nil, fmt.Errorf("session expired")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(claims.UserID, claims.Username, claims.SessionID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", claims.UserID).Msg("Failed to generate access token.")
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	s.logger.Info().Str("user_id", claims.UserID).Str("session_id", claims.SessionID).Msg("Token refreshed successfully.")

	return &models.RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(s.config.AccessExpiry),
	}, nil
}

// Logout invalidates the user session
func (s *AuthService) Logout(ctx context.Context, req *models.LogoutRequest) (*models.LogoutResponse, error) {
	if req.RefreshToken != "" {
		// Parse refresh token to get session ID
		claims, err := s.validateToken(req.RefreshToken, models.TokenTypeRefresh)
		if err == nil {
			// Delete session
			if err := s.sessionStore.Delete(ctx, claims.SessionID); err != nil {
				s.logger.Error().Err(err).Str("session_id", claims.SessionID).Msg("Failed to delete session.")
			} else {
				s.logger.Info().Str("user_id", claims.UserID).Str("session_id", claims.SessionID).Msg("User logged out successfully.")
			}
		}
	}

	return &models.LogoutResponse{
		Message: "Logged out successfully",
	}, nil
}

// GetAuthenticatedUser returns the authenticated user information
func (s *AuthService) GetAuthenticatedUser(ctx context.Context, userID string) (*models.AuthUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get authenticated user.")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &models.AuthUserResponse{
		User: user.ToResponse(),
	}, nil
}

// ValidateAccessToken validates an access token and returns claims
func (s *AuthService) ValidateAccessToken(tokenString string) (*models.Claims, error) {
	return s.validateToken(tokenString, models.TokenTypeAccess)
}

// generateAccessToken generates a new access token
func (s *AuthService) generateAccessToken(userID, username, sessionID string) (string, error) {
	claims := &models.Claims{
		UserID:    userID,
		Username:  username,
		SessionID: sessionID,
		Type:      models.TokenTypeAccess,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":    claims.UserID,
		"username":  claims.Username,
		"sessionId": claims.SessionID,
		"type":      claims.Type,
		"iss":       s.config.Issuer,
		"exp":       time.Now().Add(s.config.AccessExpiry).Unix(),
		"iat":       time.Now().Unix(),
	})

	return token.SignedString([]byte(s.config.Secret))
}

// generateRefreshToken generates a new refresh token
func (s *AuthService) generateRefreshToken(userID, username, sessionID string) (string, error) {
	claims := &models.Claims{
		UserID:    userID,
		Username:  username,
		SessionID: sessionID,
		Type:      models.TokenTypeRefresh,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":    claims.UserID,
		"username":  claims.Username,
		"sessionId": claims.SessionID,
		"type":      claims.Type,
		"iss":       s.config.Issuer,
		"exp":       time.Now().Add(s.config.RefreshExpiry).Unix(),
		"iat":       time.Now().Unix(),
	})

	return token.SignedString([]byte(s.config.Secret))
}

// validateToken validates a JWT token and returns claims
func (s *AuthService) validateToken(tokenString, expectedType string) (*models.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate token type
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != expectedType {
		return nil, fmt.Errorf("invalid token type")
	}

	// Extract claims
	userID, _ := claims["userId"].(string)
	username, _ := claims["username"].(string)
	sessionID, _ := claims["sessionId"].(string)

	if userID == "" || username == "" || sessionID == "" {
		return nil, fmt.Errorf("missing required claims")
	}

	return &models.Claims{
		UserID:    userID,
		Username:  username,
		SessionID: sessionID,
		Type:      tokenType,
	}, nil
}

// hashPassword hashes a password using bcrypt
func (s *AuthService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against its hash
func (s *AuthService) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// SetBcryptCost sets the bcrypt cost (useful for testing)
func (s *AuthService) SetBcryptCost(cost int) {
	s.bcryptCost = cost
}
