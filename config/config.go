package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server                   ServerConfig
	Database                 DatabaseConfig
	JWT                      JWTConfig
	Logging                  LoggingConfig
	CORS                     CORSConfig
	RateLimit                RateLimitConfig
	PipelineDefaultsEnabled  bool
}

// ServerConfig contains server-related settings
type ServerConfig struct {
	Port    string
	GinMode string
	Env     string
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	User            string
	Password        string
	Host            string
	Port            string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig contains JWT token settings
type JWTConfig struct {
	Secret string
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level string
}

// CORSConfig contains CORS settings
type CORSConfig struct {
	AllowedOrigins []string
	MaxAgeHours    int
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	RequestsPerMinute int
}

const (
	// Default values
	defaultServerPort        = "8081"
	defaultDBMaxOpenConns    = 25
	defaultDBMaxIdleConns    = 5
	defaultDBConnMaxLifetime = 5 * time.Minute
	defaultJWTSecret         = "dev-secret"
	defaultCORSMaxAge        = 12
	defaultRateLimit         = 60
)

var globalConfig *Config

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Attempt to load .env file; ignore if not present
	_ = godotenv.Load()

	cfg := &Config{
		Server:                  loadServerConfig(),
		Database:                loadDatabaseConfig(),
		JWT:                     loadJWTConfig(),
		Logging:                 loadLoggingConfig(),
		CORS:                    loadCORSConfig(),
		RateLimit:               loadRateLimitConfig(),
		PipelineDefaultsEnabled: getEnvAsBool("PIPELINE_DEFAULTS_ENABLED", true),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// Get returns the global config instance
func Get() *Config {
	return globalConfig
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Port:    getEnvOrDefault("SERVER_PORT", defaultServerPort),
		GinMode: getEnvOrDefault("GIN_MODE", ""),
		Env:     strings.ToLower(getEnvOrDefault("ENV", "local")),
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		User:            os.Getenv("DB_USER"),
		Password:        os.Getenv("DB_PASSWORD"),
		Host:            os.Getenv("DB_HOST"),
		Port:            os.Getenv("DB_PORT"),
		Name:            os.Getenv("DB_NAME"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", defaultDBMaxOpenConns),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", defaultDBMaxIdleConns),
		ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", defaultDBConnMaxLifetime),
	}
}

func loadJWTConfig() JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = defaultJWTSecret
	}

	return JWTConfig{
		Secret: secret,
	}
}

func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level: strings.ToLower(getEnvOrDefault("LOG_LEVEL", "")),
	}
}

func loadCORSConfig() CORSConfig {
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}

	origins := strings.Split(allowedOrigins, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	return CORSConfig{
		AllowedOrigins: origins,
		MaxAgeHours:    getEnvAsInt("CORS_MAX_AGE_HOURS", defaultCORSMaxAge),
	}
}

func loadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: getEnvAsInt("RATE_LIMIT_PER_MINUTE", defaultRateLimit),
	}
}

func (c *Config) validate() error {
	if c.Database.User == "" || c.Database.Password == "" ||
		c.Database.Host == "" || c.Database.Port == "" ||
		c.Database.Name == "" {
		return fmt.Errorf("database configuration incomplete: ensure DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, and DB_NAME are set")
	}
	return nil
}

// DSN returns the MySQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
