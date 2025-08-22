package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Log       LogConfig       `mapstructure:"log"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	Environment  string        `mapstructure:"environment"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	PostgresURL  string `mapstructure:"postgres_url"`
	MongoURL     string `mapstructure:"mongo_url"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
	Issuer        string        `mapstructure:"issuer"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int           `mapstructure:"requests"`
	Window   time.Duration `mapstructure:"window"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Set up Viper
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults()

	// Bind environment variables explicitly
	bindEnvVars()

	// Create config struct
	var config Config

	// Unmarshal configuration
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// bindEnvVars explicitly binds environment variables to Viper keys
func bindEnvVars() {
	// Server configuration
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	viper.BindEnv("server.environment", "SERVER_ENVIRONMENT")

	// Database configuration
	viper.BindEnv("database.driver", "DATABASE_DRIVER")
	viper.BindEnv("database.postgres_url", "DATABASE_POSTGRES_URL")
	viper.BindEnv("database.mongo_url", "DATABASE_MONGO_URL")
	viper.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")

	// Redis configuration
	viper.BindEnv("redis.url", "REDIS_URL")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	// JWT configuration
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.access_expiry", "JWT_ACCESS_EXPIRY")
	viper.BindEnv("jwt.refresh_expiry", "JWT_REFRESH_EXPIRY")
	viper.BindEnv("jwt.issuer", "JWT_ISSUER")

	// Rate limit configuration
	viper.BindEnv("rate_limit.requests", "RATE_LIMIT_REQUESTS")
	viper.BindEnv("rate_limit.window", "RATE_LIMIT_WINDOW")

	// Log configuration
	viper.BindEnv("log.level", "LOG_LEVEL")
	viper.BindEnv("log.format", "LOG_FORMAT")
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 9000)
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.environment", "development")

	// Database defaults
	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)

	// Redis defaults
	viper.SetDefault("redis.url", "redis://localhost:6379/0")
	viper.SetDefault("redis.db", 0)

	// JWT defaults
	viper.SetDefault("jwt.access_expiry", "15m")
	viper.SetDefault("jwt.refresh_expiry", "168h")
	viper.SetDefault("jwt.issuer", "go-fiber")

	// Rate limit defaults
	viper.SetDefault("rate_limit.requests", 100)
	viper.SetDefault("rate_limit.window", "1m")

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
}

// validate validates the configuration
func validate(config *Config) error {
	// Validate server configuration
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validate database configuration
	if config.Database.Driver != "postgres" && config.Database.Driver != "mongodb" {
		return fmt.Errorf("unsupported database driver: %s", config.Database.Driver)
	}

	if config.Database.Driver == "postgres" && config.Database.PostgresURL == "" {
		return fmt.Errorf("postgres_url is required when using postgres driver")
	}

	if config.Database.Driver == "mongodb" && config.Database.MongoURL == "" {
		return fmt.Errorf("mongo_url is required when using mongodb driver")
	}

	// Validate JWT configuration
	if config.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}

	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("jwt secret must be at least 32 characters long")
	}

	// Validate Redis configuration
	if config.Redis.URL == "" {
		return fmt.Errorf("redis url is required")
	}

	return nil
}

// GetAddress returns the server address in host:port format
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}
