// internal/config/config.go
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	Port         string
	SolanaRPCURL string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		// Not an error if .env doesn't exist in production
		// as we'll use actual environment variables
	}

	return &Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/blockhawk?sslmode=disable"),
		Port:         getEnv("PORT", "8080"),
		SolanaRPCURL: getEnv("SOLANA_RPC_URL", "https://api.mainnet-beta.solana.com"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
