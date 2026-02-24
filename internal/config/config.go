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

	// AI providers — keys for all supported providers; AIProvider selects
	// the default on startup. Switchable at runtime from admin Settings.
	AIProvider string // Default active: "openai", "gemini", "claude", "mistral"

	// Per-provider credentials
	OpenAIKey     string
	OpenAIModel   string
	OpenAIBaseURL string

	GeminiKey     string
	GeminiModel   string
	GeminiBaseURL string

	ClaudeKey     string
	ClaudeModel   string
	ClaudeBaseURL string

	MistralKey     string
	MistralModel   string
	MistralBaseURL string
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

		AIProvider: envOrDefault("AI_PROVIDER", "gemini"),

		OpenAIKey:     os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:   envOrDefault("OPENAI_MODEL", "gpt-4o"),
		OpenAIBaseURL: envOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),

		GeminiKey:     os.Getenv("GEMINI_API_KEY"),
		GeminiModel:   envOrDefault("GEMINI_MODEL", "gemini-3.1-pro-preview"),
		GeminiBaseURL: envOrDefault("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com"),

		ClaudeKey:     os.Getenv("CLAUDE_API_KEY"),
		ClaudeModel:   envOrDefault("CLAUDE_MODEL", "claude-sonnet-4-6"),
		ClaudeBaseURL: envOrDefault("CLAUDE_BASE_URL", "https://api.anthropic.com"),

		MistralKey:     os.Getenv("MISTRAL_API_KEY"),
		MistralModel:   envOrDefault("MISTRAL_MODEL", "mistral-large-latest"),
		MistralBaseURL: envOrDefault("MISTRAL_BASE_URL", "https://api.mistral.ai"),
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
