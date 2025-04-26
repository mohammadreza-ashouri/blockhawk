package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Monitoring MonitoringConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port          string
	Host          string
	ReadTimeout   int
	WriteTimeout  int
	StaticDir     string
	TemplatesDir  string
	SessionSecret string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Type     string // "sqlite" or "postgres"
	Path     string // For SQLite
	Host     string // For PostgreSQL
	Port     string // For PostgreSQL
	User     string // For PostgreSQL
	Password string // For PostgreSQL
	Name     string // For PostgreSQL
	SSLMode  string // For PostgreSQL
}

// MonitoringConfig holds monitoring-related configuration
type MonitoringConfig struct {
	EnableSolana         bool
	EnableRipple         bool
	EnableEthereum       bool
	DefaultAlertLimit    int
	HistoryRetentionDays int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	config := &Config{
		Server: ServerConfig{
			Port:          getEnv("PORT", "8080"),
			Host:          getEnv("HOST", "0.0.0.0"),
			ReadTimeout:   getEnvAsInt("READ_TIMEOUT", 15),
			WriteTimeout:  getEnvAsInt("WRITE_TIMEOUT", 15),
			StaticDir:     getEnv("STATIC_DIR", "./web/static"),
			TemplatesDir:  getEnv("TEMPLATES_DIR", "./web/templates"),
			SessionSecret: getEnv("SESSION_SECRET", "blockhawk-secret-key"),
		},
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "sqlite"),
			Path:     getEnv("DB_PATH", "./data/blockhawk.db"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "blockhawk"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Monitoring: MonitoringConfig{
			EnableSolana:         getEnvAsBool("ENABLE_SOLANA", true),
			EnableRipple:         getEnvAsBool("ENABLE_RIPPLE", true),
			EnableEthereum:       getEnvAsBool("ENABLE_ETHEREUM", false),
			DefaultAlertLimit:    getEnvAsInt("DEFAULT_ALERT_LIMIT", 50),
			HistoryRetentionDays: getEnvAsInt("HISTORY_RETENTION_DAYS", 30),
		},
	}

	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as an integer
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsBool gets an environment variable as a boolean
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	// Convert string to lowercase for comparison
	valueStr = strings.ToLower(valueStr)

	return valueStr == "true" || valueStr == "yes" || valueStr == "1"
}
