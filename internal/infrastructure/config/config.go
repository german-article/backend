package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	ProjectID       string
	ApplicationName string
	TelegramToken   string
	GCPEnabled      bool
	LogLevel        int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		ProjectID:       getEnv("PROJECT_ID", "german-article-bot"),
		ApplicationName: getEnv("APPLICATION_NAME", "article-bot"),
		TelegramToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		GCPEnabled:      getEnv("GCP_ENABLED", "true") == "true",
		LogLevel:        100, // Default log level
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
