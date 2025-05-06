// Add this if you don't already have a config.go file

package config

import (
	"os"
)

// AppConfig holds application configuration
type AppConfig struct {
	FrontendURL string
}

// GetConfig returns the application configuration
func GetConfig() AppConfig {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	return AppConfig{
		FrontendURL: frontendURL,
	}
}
