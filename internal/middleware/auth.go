package middleware

import (
	"strings"

	"go-fiber/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(authService *services.AuthService, logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Warn().Str("path", c.Path()).Msg("Missing Authorization header.")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Missing authorization header",
			})
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn().Str("path", c.Path()).Msg("Invalid Authorization header format.")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			logger.Warn().Str("path", c.Path()).Msg("Empty token.")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Empty token",
			})
		}

		// Validate token
		claims, err := authService.ValidateAccessToken(token)
		if err != nil {
			logger.Warn().Err(err).Str("path", c.Path()).Msg("Invalid token.")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid token",
			})
		}

		// Store user information in context
		c.Locals("userID", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("sessionID", claims.SessionID)

		logger.Debug().
			Str("user_id", claims.UserID).
			Str("username", claims.Username).
			Str("path", c.Path()).
			Msg("User authenticated successfully.")

		return c.Next()
	}
}

// OptionalAuthMiddleware creates optional JWT authentication middleware
// This middleware will set user context if a valid token is provided, but won't fail if no token is present
func OptionalAuthMiddleware(authService *services.AuthService, logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			return c.Next()
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Invalid format, continue without authentication
			return c.Next()
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			// Empty token, continue without authentication
			return c.Next()
		}

		// Validate token
		claims, err := authService.ValidateAccessToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			logger.Debug().Err(err).Str("path", c.Path()).Msg("Invalid optional token.")
			return c.Next()
		}

		// Store user information in context
		c.Locals("userID", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("sessionID", claims.SessionID)

		logger.Debug().
			Str("user_id", claims.UserID).
			Str("username", claims.Username).
			Str("path", c.Path()).
			Msg("User authenticated via optional middleware.")

		return c.Next()
	}
}

// GetUserID extracts user ID from Fiber context
func GetUserID(c *fiber.Ctx) string {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUsername extracts username from Fiber context
func GetUsername(c *fiber.Ctx) string {
	username, ok := c.Locals("username").(string)
	if !ok {
		return ""
	}
	return username
}

// GetSessionID extracts session ID from Fiber context
func GetSessionID(c *fiber.Ctx) string {
	sessionID, ok := c.Locals("sessionID").(string)
	if !ok {
		return ""
	}
	return sessionID
}

// RequireAuth ensures that the user is authenticated
func RequireAuth(c *fiber.Ctx) error {
	userID := GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}
	return nil
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(c *fiber.Ctx) bool {
	return GetUserID(c) != ""
}
