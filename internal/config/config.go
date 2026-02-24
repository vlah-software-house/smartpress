// Package config handles application configuration loading from environment
// variables. It provides a centralized Config struct used across the application.
package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration values loaded from the environment.
type Config struct {
	// Server settings
	Host string
	Port string
	Env  string // "development", "production", "testing"

	// PostgreSQL connection
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Valkey (Redis-compatible cache)
	ValkeyHost     string
	ValkeyPort     string
	ValkeyPassword string

	// AI provider settings
	AIProvider string // "openai", "gemini", "claude"
	AIAPIKey   string
	AIModel    string
}

// Load reads configuration from environment variables, applying defaults
// for development where appropriate. Returns an error if critical values
// are missing in production mode.
func Load() (*Config, error) {
	cfg := &Config{
		Host: envOrDefault("APP_HOST", "0.0.0.0"),
		Port: envOrDefault("APP_PORT", "8080"),
		Env:  envOrDefault("APP_ENV", "development"),

		DBHost:     envOrDefault("POSTGRES_HOST", "localhost"),
		DBPort:     envOrDefault("POSTGRES_PORT", "5432"),
		DBUser:     envOrDefault("POSTGRES_USER", "smartpress"),
		DBPassword: envOrDefault("POSTGRES_PASSWORD", "changeme"),
		DBName:     envOrDefault("POSTGRES_DB", "smartpress"),

		ValkeyHost:     envOrDefault("VALKEY_HOST", "localhost"),
		ValkeyPort:     envOrDefault("VALKEY_PORT", "6379"),
		ValkeyPassword: os.Getenv("VALKEY_PASSWORD"),

		AIProvider: os.Getenv("AI_PROVIDER"),
		AIAPIKey:   os.Getenv("AI_API_KEY"),
		AIModel:    os.Getenv("AI_MODEL"),
	}

	if cfg.Env == "production" {
		if cfg.DBPassword == "changeme" {
			return nil, fmt.Errorf("POSTGRES_PASSWORD must be set in production")
		}
	}

	return cfg, nil
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// Addr returns the server listen address (host:port).
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// IsDev returns true if the application is running in development mode.
func (c *Config) IsDev() bool {
	return c.Env == "development"
}

// envOrDefault reads an environment variable, returning a fallback if unset or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
