package server

import (
	"go-fiber/internal/middleware"

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// setupRoutes configures all application routes
func (s *Server) setupRoutes() {
	// Swagger documentation
	s.app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Health check routes
	s.healthHandler.RegisterRoutes(s.app)

	// API routes
	api := s.app.Group("/api/v1")

	// Auth routes (no middleware required)
	auth := api.Group("/auth")
	auth.Post("/register", s.authHandler.Register)
	auth.Post("/login", s.authHandler.Login)
	auth.Post("/refresh", s.authHandler.RefreshToken)
	auth.Post("/logout", middleware.AuthMiddleware(s.authService, s.logger), s.authHandler.Logout)
	auth.Get("/me", middleware.AuthMiddleware(s.authService, s.logger), s.authHandler.Me)

	// Protected routes
	authMiddleware := middleware.AuthMiddleware(s.authService, s.logger)

	// Todo routes
	s.todoHandler.RegisterRoutes(api, authMiddleware)

	s.logger.Info().Msg("Routes setup completed.")
}
