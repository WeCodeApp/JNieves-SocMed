package config

import (
	"log"
	"os"

	"golang.org/x/oauth2"
)

func MicrosoftOAuthConfig() *oauth2.Config {
	clientID := os.Getenv("MICROSOFT_CLIENT_ID")
	clientSecret := os.Getenv("MICROSOFT_CLIENT_SECRET")
	redirectURL := os.Getenv("MICROSOFT_REDIRECT_URL")

	// Log values for debugging (remove in production)
	log.Printf("OAuth Config - ClientID: %s, RedirectURL: %s",
		clientID, redirectURL)

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		log.Fatal("Missing required OAuth environment variables")
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
		},
		RedirectURL: redirectURL,
		Scopes: []string{
			"openid",
			"profile",
			"email",
			"User.Read",
			"offline_access",
		},
	}
}
