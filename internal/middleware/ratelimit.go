package middleware

import (
	"time"

	"go-fiber/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimit creates a rate limiting middleware
func RateLimit(cfg config.RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Requests,
		Expiration: cfg.Window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too Many Requests",
				"message":     "Rate limit exceeded. Please try again later.",
				"retry_after": cfg.Window.Seconds(),
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		LimiterMiddleware:      limiter.SlidingWindow{},
	})
}

// AuthRateLimit creates a stricter rate limiting middleware for authentication endpoints
func AuthRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5, // 5 requests per minute for auth endpoints
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too Many Requests",
				"message":     "Too many authentication attempts. Please try again later.",
				"retry_after": 60,
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		LimiterMiddleware:      limiter.SlidingWindow{},
	})
}

// APIRateLimit creates a rate limiting middleware for API endpoints
func APIRateLimit(cfg config.RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Requests * 2, // More lenient for API endpoints
		Expiration: cfg.Window,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use user ID if authenticated, otherwise IP
			userID := c.Locals("userID")
			if userID != nil {
				return "user:" + userID.(string)
			}
			return "ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too Many Requests",
				"message":     "API rate limit exceeded. Please try again later.",
				"retry_after": cfg.Window.Seconds(),
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		LimiterMiddleware:      limiter.SlidingWindow{},
	})
}
