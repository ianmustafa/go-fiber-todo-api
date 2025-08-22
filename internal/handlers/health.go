package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	pgDB    *pgxpool.Pool
	mongoDB *mongo.Database
	redis   redis.Cmdable
	logger  zerolog.Logger
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]ServiceInfo `json:"services"`
}

// ServiceInfo represents the status of a service
type ServiceInfo struct {
	Status       string `json:"status"`
	ResponseTime string `json:"responseTime"`
	Error        string `json:"error,omitempty"`
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(pgDB *pgxpool.Pool, mongoDB *mongo.Database, redis redis.Cmdable, logger zerolog.Logger) *HealthHandler {
	return &HealthHandler{
		pgDB:    pgDB,
		mongoDB: mongoDB,
		redis:   redis,
		logger:  logger,
	}
}

// HealthCheck handles basic health check
// @Summary Health check
// @Description Check if the service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	response := &HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0", // This could be injected from build info
		Services:  make(map[string]ServiceInfo),
	}

	// Check PostgreSQL
	if h.pgDB != nil {
		start := time.Now()
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		err := h.pgDB.Ping(ctx)
		responseTime := time.Since(start)

		if err != nil {
			response.Services["postgresql"] = ServiceInfo{
				Status:       "unhealthy",
				ResponseTime: responseTime.String(),
				Error:        err.Error(),
			}
			response.Status = "degraded"
			h.logger.Error().Err(err).Msg("PostgreSQL health check failed")
		} else {
			response.Services["postgresql"] = ServiceInfo{
				Status:       "healthy",
				ResponseTime: responseTime.String(),
			}
		}
	}

	// Check MongoDB
	if h.mongoDB != nil {
		start := time.Now()
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		err := h.mongoDB.Client().Ping(ctx, readpref.Primary())
		responseTime := time.Since(start)

		if err != nil {
			response.Services["mongodb"] = ServiceInfo{
				Status:       "unhealthy",
				ResponseTime: responseTime.String(),
				Error:        err.Error(),
			}
			response.Status = "degraded"
			h.logger.Error().Err(err).Msg("MongoDB health check failed")
		} else {
			response.Services["mongodb"] = ServiceInfo{
				Status:       "healthy",
				ResponseTime: responseTime.String(),
			}
		}
	}

	// Check Redis
	if h.redis != nil {
		start := time.Now()
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		err := h.redis.Ping(ctx).Err()
		responseTime := time.Since(start)

		if err != nil {
			response.Services["redis"] = ServiceInfo{
				Status:       "unhealthy",
				ResponseTime: responseTime.String(),
				Error:        err.Error(),
			}
			response.Status = "degraded"
			h.logger.Error().Err(err).Msg("Redis health check failed")
		} else {
			response.Services["redis"] = ServiceInfo{
				Status:       "healthy",
				ResponseTime: responseTime.String(),
			}
		}
	}

	// Determine overall status
	if response.Status == "healthy" {
		return c.JSON(response)
	} else {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}
}

// ReadinessCheck handles readiness check
// @Summary Readiness check
// @Description Check if the service is ready to serve requests
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(c *fiber.Ctx) error {
	response := &HealthResponse{
		Status:    "ready",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services:  make(map[string]ServiceInfo),
	}

	allHealthy := true

	// Check all critical services for readiness
	if h.pgDB != nil {
		start := time.Now()
		ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
		defer cancel()

		err := h.pgDB.Ping(ctx)
		responseTime := time.Since(start)

		if err != nil {
			response.Services["postgresql"] = ServiceInfo{
				Status:       "not_ready",
				ResponseTime: responseTime.String(),
				Error:        err.Error(),
			}
			allHealthy = false
		} else {
			response.Services["postgresql"] = ServiceInfo{
				Status:       "ready",
				ResponseTime: responseTime.String(),
			}
		}
	}

	if h.mongoDB != nil {
		start := time.Now()
		ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
		defer cancel()

		err := h.mongoDB.Client().Ping(ctx, readpref.Primary())
		responseTime := time.Since(start)

		if err != nil {
			response.Services["mongodb"] = ServiceInfo{
				Status:       "not_ready",
				ResponseTime: responseTime.String(),
				Error:        err.Error(),
			}
			allHealthy = false
		} else {
			response.Services["mongodb"] = ServiceInfo{
				Status:       "ready",
				ResponseTime: responseTime.String(),
			}
		}
	}

	if h.redis != nil {
		start := time.Now()
		ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
		defer cancel()

		err := h.redis.Ping(ctx).Err()
		responseTime := time.Since(start)

		if err != nil {
			response.Services["redis"] = ServiceInfo{
				Status:       "not_ready",
				ResponseTime: responseTime.String(),
				Error:        err.Error(),
			}
			allHealthy = false
		} else {
			response.Services["redis"] = ServiceInfo{
				Status:       "ready",
				ResponseTime: responseTime.String(),
			}
		}
	}

	if !allHealthy {
		response.Status = "not_ready"
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	return c.JSON(response)
}

// LivenessCheck handles liveness check
// @Summary Liveness check
// @Description Check if the service is alive
// @Tags health
// @Produce json
// @Success 200 {object} models.MessageResponse
// @Router /live [get]
func (h *HealthHandler) LivenessCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "alive",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/health", h.HealthCheck)
	router.Get("/ready", h.ReadinessCheck)
	router.Get("/live", h.LivenessCheck)
}
