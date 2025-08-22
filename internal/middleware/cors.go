package middleware

import (
	"go-fiber/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS creates a CORS middleware with configuration
func CORS(cfg *config.Config) fiber.Handler {
	corsConfig := cors.Config{
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: false,
		MaxAge:           300,
	}

	// Configure origins based on environment
	if cfg.IsDevelopment() {
		corsConfig.AllowOrigins = "*"
	} else {
		// In production, specify allowed origins
		corsConfig.AllowOrigins = "https://yourdomain.com,https://www.yourdomain.com"
	}

	return cors.New(corsConfig)
}

// CORSWithOrigins creates a CORS middleware with specific origins
func CORSWithOrigins(origins string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
		MaxAge:           300,
	})
}

// CORSStrict creates a strict CORS middleware for production
func CORSStrict(allowedOrigins []string) fiber.Handler {
	origins := ""
	for i, origin := range allowedOrigins {
		if i > 0 {
			origins += ","
		}
		origins += origin
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})
}
