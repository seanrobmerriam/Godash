package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database configuration (for future use)
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

// SessionConfig holds session configuration
type SessionConfig struct {
	SecretKey string
	MaxAge    int
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "sqlite"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Name:     getEnv("DB_NAME", "godash.db"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
		},
		Session: SessionConfig{
			SecretKey: getEnv("SESSION_SECRET", "change-this-secret-key-in-production"),
			MaxAge:    getEnvAsInt("SESSION_MAX_AGE", 86400), // 24 hours
		},
	}
}

// Helper functions
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		} else {
			log.Printf("Warning: Invalid integer value for %s: %s, using default: %d", key, value, defaultVal)
		}
	}
	return defaultVal
}