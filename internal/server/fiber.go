package server

import (
	"github.com/gofiber/fiber/v2"
)

// setupFiberApp creates and configures the Fiber application
func (s *Server) setupFiberApp() {
	s.app = fiber.New(fiber.Config{
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		ErrorHandler: s.customErrorHandler(),
		AppName:      "Go Fiber Todo API v1.0.0",
	})
}

// customErrorHandler handles errors globally
func (s *Server) customErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		s.logger.Error().
			Err(err).
			Int("status", code).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Msg("Request error.")

		return c.Status(code).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
	}
}
