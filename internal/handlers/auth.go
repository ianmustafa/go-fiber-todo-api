package handlers

import (
	"go-fiber/internal/middleware"
	"go-fiber/internal/models"
	"go-fiber/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *services.AuthService
	validator   *validator.Validate
	logger      zerolog.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService, validator *validator.Validate, logger zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
		logger:      logger,
	}
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	auth := router.Group("/auth")

	// Public routes
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/login/email", h.LoginByEmail)
	auth.Post("/refresh", h.RefreshToken)
	auth.Post("/logout", h.Logout)

	// Protected routes
	auth.Get("/me", authMiddleware, h.Me)
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration request"
// @Success 201 {object} models.RegisterResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse registration request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error().Err(err).Msg("Registration request validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation Error",
			"message": "Invalid input data",
			"details": err.Error(),
		})
	}

	// Register user
	response, err := h.authService.Register(c.Context(), &req)
	if err != nil {
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   "Conflict",
				"message": err.Error(),
			})
		}
		h.logger.Error().Err(err).Msg("Failed to register user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to register user",
		})
	}

	h.logger.Info().Str("username", req.Username).Msg("User registered successfully")
	return c.Status(fiber.StatusCreated).JSON(response)
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse login request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error().Err(err).Msg("Login request validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation Error",
			"message": "Invalid input data",
			"details": err.Error(),
		})
	}

	// Login user
	response, err := h.authService.Login(c.Context(), &req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid credentials",
			})
		}
		h.logger.Error().Err(err).Msg("Failed to login user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to login user",
		})
	}

	h.logger.Info().Str("username", req.Username).Msg("User logged in successfully")
	return c.JSON(response)
}

// LoginByEmail handles user login by email
// @Summary Login user by email
// @Description Authenticate user by email and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginByEmailRequest true "Login by email request"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/login/email [post]
func (h *AuthHandler) LoginByEmail(c *fiber.Ctx) error {
	var req models.LoginByEmailRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse login by email request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error().Err(err).Msg("Login by email request validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation Error",
			"message": "Invalid input data",
			"details": err.Error(),
		})
	}

	// Login user by email
	response, err := h.authService.LoginByEmail(c.Context(), &req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid credentials",
			})
		}
		h.logger.Error().Err(err).Msg("Failed to login user by email")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to login user",
		})
	}

	h.logger.Info().Str("email", req.Email).Msg("User logged in by email successfully")
	return c.JSON(response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} models.RefreshTokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req models.RefreshTokenRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse refresh token request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error().Err(err).Msg("Refresh token request validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation Error",
			"message": "Invalid input data",
			"details": err.Error(),
		})
	}

	// Refresh token
	response, err := h.authService.RefreshToken(c.Context(), &req)
	if err != nil {
		if err.Error() == "invalid refresh token" || err.Error() == "invalid session" || err.Error() == "session expired" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": err.Error(),
			})
		}
		h.logger.Error().Err(err).Msg("Failed to refresh token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to refresh token",
		})
	}

	h.logger.Info().Msg("Token refreshed successfully")
	return c.JSON(response)
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate user session
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LogoutRequest true "Logout request"
// @Success 200 {object} models.LogoutResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req models.LogoutRequest

	// Parse request body (optional refresh token)
	if err := c.BodyParser(&req); err != nil {
		// If parsing fails, continue with empty request (logout without refresh token)
		req = models.LogoutRequest{}
	}

	// Logout user
	response, err := h.authService.Logout(c.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to logout user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to logout user",
		})
	}

	h.logger.Info().Msg("User logged out successfully")
	return c.JSON(response)
}

// Me handles getting current user information
// @Summary Get current user
// @Description Get authenticated user information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.AuthUserResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
	}

	// Get user information
	response, err := h.authService.GetAuthenticatedUser(c.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get authenticated user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get user information",
		})
	}

	return c.JSON(response)
}
