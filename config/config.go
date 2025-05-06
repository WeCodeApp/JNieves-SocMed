package config

import (
	"os"
)

type AppConfig struct {
	FrontendURL string
}

func GetConfig() AppConfig {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	return AppConfig{
		FrontendURL: frontendURL,
	}
}
