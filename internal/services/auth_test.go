package services

import (
	"context"
	"testing"
	"time"

	"go-fiber/internal/config"
	"go-fiber/internal/mocks"
	"go-fiber/internal/models"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockSessionStore := new(mocks.MockSessionStore)
	logger := zerolog.Nop()
	jwtConfig := &config.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  time.Hour,
		RefreshExpiry: 24 * time.Hour,
		Issuer:        "test-issuer",
	}

	authService := NewAuthService(mockUserRepo, mockSessionStore, jwtConfig, logger)
	authService.SetBcryptCost(bcrypt.MinCost) // Use minimum cost for testing

	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		// Arrange
		req := &models.RegisterRequest{
			Username: "testuser",
			Password: "password123",
			Email:    "test@example.com",
		}

		expectedUser := &models.User{
			ID:       "test-id",
			Username: "testuser",
			Email:    "test@example.com",
		}

		mockUserRepo.On("ExistsByUsername", ctx, "testuser").Return(false, nil)
		mockUserRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(expectedUser, nil)

		// Act
		result, err := authService.Register(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "testuser", result.User.Username)
		assert.Equal(t, "test@example.com", result.User.Email)
		assert.Equal(t, "User registered successfully", result.Message)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("username already exists", func(t *testing.T) {
		// Arrange
		req := &models.RegisterRequest{
			Username: "existinguser",
			Password: "password123",
		}

		mockUserRepo.On("ExistsByUsername", ctx, "existinguser").Return(true, nil)

		// Act
		result, err := authService.Register(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "username already exists")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		// Arrange
		req := &models.RegisterRequest{
			Username: "newuser",
			Password: "password123",
			Email:    "existing@example.com",
		}

		mockUserRepo.On("ExistsByUsername", ctx, "newuser").Return(false, nil)
		mockUserRepo.On("ExistsByEmail", ctx, "existing@example.com").Return(true, nil)

		// Act
		result, err := authService.Register(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "email already exists")

		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockSessionStore := new(mocks.MockSessionStore)
	logger := zerolog.Nop()
	jwtConfig := &config.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  time.Hour,
		RefreshExpiry: 24 * time.Hour,
		Issuer:        "test-issuer",
	}

	authService := NewAuthService(mockUserRepo, mockSessionStore, jwtConfig, logger)
	authService.SetBcryptCost(bcrypt.MinCost)

	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		// Arrange
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

		req := &models.LoginRequest{
			Username: "testuser",
			Password: password,
		}

		user := &models.User{
			ID:       "test-id",
			Username: "testuser",
			Password: string(hashedPassword),
			Email:    "test@example.com",
		}

		mockUserRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)
		mockSessionStore.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*models.Session"), mock.AnythingOfType("time.Duration")).Return(nil)

		// Act
		result, err := authService.Login(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.Equal(t, "testuser", result.User.Username)

		mockUserRepo.AssertExpectations(t)
		mockSessionStore.AssertExpectations(t)
	})

	t.Run("invalid username", func(t *testing.T) {
		// Arrange
		req := &models.LoginRequest{
			Username: "nonexistent",
			Password: "password123",
		}

		mockUserRepo.On("GetByUsername", ctx, "nonexistent").Return(nil, assert.AnError)

		// Act
		result, err := authService.Login(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid credentials")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		// Arrange
		correctPassword := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.MinCost)

		req := &models.LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}

		user := &models.User{
			ID:       "test-id",
			Username: "testuser",
			Password: string(hashedPassword),
		}

		mockUserRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)

		// Act
		result, err := authService.Login(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid credentials")

		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_ValidateAccessToken(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockSessionStore := new(mocks.MockSessionStore)
	logger := zerolog.Nop()
	jwtConfig := &config.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  time.Hour,
		RefreshExpiry: 24 * time.Hour,
		Issuer:        "test-issuer",
	}

	authService := NewAuthService(mockUserRepo, mockSessionStore, jwtConfig, logger)

	t.Run("valid token", func(t *testing.T) {
		// Arrange - Generate a valid token
		token, err := authService.generateAccessToken("user-id", "testuser", "session-id")
		assert.NoError(t, err)

		// Act
		claims, err := authService.ValidateAccessToken(token)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "user-id", claims.UserID)
		assert.Equal(t, "testuser", claims.Username)
		assert.Equal(t, "session-id", claims.SessionID)
		assert.Equal(t, models.TokenTypeAccess, claims.Type)
	})

	t.Run("invalid token", func(t *testing.T) {
		// Act
		claims, err := authService.ValidateAccessToken("invalid-token")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("wrong token type", func(t *testing.T) {
		// Arrange - Generate a refresh token instead of access token
		token, err := authService.generateRefreshToken("user-id", "testuser", "session-id")
		assert.NoError(t, err)

		// Act
		claims, err := authService.ValidateAccessToken(token)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "invalid token type")
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockSessionStore := new(mocks.MockSessionStore)
	logger := zerolog.Nop()
	jwtConfig := &config.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  time.Hour,
		RefreshExpiry: 24 * time.Hour,
		Issuer:        "test-issuer",
	}

	authService := NewAuthService(mockUserRepo, mockSessionStore, jwtConfig, logger)
	ctx := context.Background()

	t.Run("successful token refresh", func(t *testing.T) {
		// Arrange
		refreshToken, err := authService.generateRefreshToken("user-id", "testuser", "session-id")
		assert.NoError(t, err)

		req := &models.RefreshTokenRequest{
			RefreshToken: refreshToken,
		}

		session := &models.Session{
			ID:        "session-id",
			UserID:    "user-id",
			IsActive:  true,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockSessionStore.On("Get", ctx, "session-id").Return(session, nil)

		// Act
		result, err := authService.RefreshToken(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)

		mockSessionStore.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		// Arrange
		req := &models.RefreshTokenRequest{
			RefreshToken: "invalid-token",
		}

		// Act
		result, err := authService.RefreshToken(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid refresh token")
	})

	t.Run("expired session", func(t *testing.T) {
		// Arrange
		refreshToken, err := authService.generateRefreshToken("user-id", "testuser", "session-id")
		assert.NoError(t, err)

		req := &models.RefreshTokenRequest{
			RefreshToken: refreshToken,
		}

		session := &models.Session{
			ID:        "session-id",
			UserID:    "user-id",
			IsActive:  true,
			ExpiresAt: time.Now().Add(-time.Hour), // Expired
		}

		mockSessionStore.On("Get", ctx, "session-id").Return(session, nil)

		// Act
		result, err := authService.RefreshToken(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "session expired")

		mockSessionStore.AssertExpectations(t)
	})
}
