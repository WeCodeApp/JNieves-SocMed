package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"go-rest-api/config"
	"go-rest-api/database"
	"go-rest-api/internal/models"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type AuthController struct {
	store *session.Store
}

func NewAuthController(store *session.Store) *AuthController {
	return &AuthController{
		store: store,
	}
}

// MicrosoftLogin initiates Microsoft OAuth flow
func (ac *AuthController) MicrosoftLogin(c *fiber.Ctx) error {
	// Generate random state
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate state",
		})
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Store state in session
	sess, err := ac.store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get session",
		})
	}
	sess.Set("oauth_state", state)
	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save session",
		})
	}

	// Get the OAuth config
	oauthConfig := config.MicrosoftOAuthConfig()

	// Log the authorization URL for debugging
	authURL := oauthConfig.AuthCodeURL(state)
	log.Printf("Redirecting to: %s", authURL)

	// Redirect to Microsoft login
	return c.Redirect(authURL)
}

// MicrosoftCallback handles the OAuth callback from Microsoft
func (ac *AuthController) MicrosoftCallback(c *fiber.Ctx) error {
	// Log all query parameters for debugging
	queryParams := c.Queries()
	log.Printf("Callback received. Query params: %v", queryParams)

	// Check for error from OAuth provider
	if errorMsg := c.Query("error"); errorMsg != "" {
		errorDesc := c.Query("error_description")
		log.Printf("OAuth Error from provider: %s - %s", errorMsg, errorDesc)
		return c.JSON(fiber.Map{
			"error":       "Authentication failed: " + errorMsg,
			"description": errorDesc,
		})
	}

	// Get authorization code from query parameters
	code := c.Query("code")
	if code == "" {
		log.Printf("Error: Missing authorization code in callback")
		return c.JSON(fiber.Map{
			"error": "Missing authorization code",
		})
	}

	// Get state from session
	sess, err := ac.store.Get(c)
	if err != nil {
		log.Printf("Session error: %v", err)
		return c.JSON(fiber.Map{
			"error": "Failed to get session: " + err.Error(),
		})
	}

	expectedState := sess.Get("oauth_state")
	receivedState := c.Query("state")

	// Verify state
	if receivedState == "" || receivedState != expectedState {
		log.Printf("State mismatch. Expected: %v, Received: %v", expectedState, receivedState)
		return c.JSON(fiber.Map{
			"error": "Invalid OAuth state",
		})
	}

	// Exchange code for token
	oauthConfig := config.MicrosoftOAuthConfig()

	// Log before token exchange
	log.Printf("Attempting to exchange code for token. Code length: %d", len(code))

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Token exchange error: %v", err)
		return c.JSON(fiber.Map{
			"error": "Failed to exchange token: " + err.Error(),
		})
	}

	// Successfully exchanged token, continue with your existing logic...
	// Removed unused context declaration

	// Get user info using Microsoft Graph API
	client := oauthConfig.Client(context.Background(), token)

	// Use the correct MS Graph API endpoint for consumer accounts
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		return c.Redirect("/login?error=profile_fetch_failed")
	}
	defer resp.Body.Close()

	userData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read user data: " + err.Error(),
		})
	}

	// Parse user data
	var userInfo map[string]interface{}
	if err := json.Unmarshal(userData, &userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse user data",
		})
	}

	// Log received user data for debugging
	log.Printf("User info received: %v", userInfo)

	// Find or create user using SQL
	email, ok := userInfo["mail"].(string)
	if !ok {
		// Fallback to userPrincipalName if mail is not available
		email, ok = userInfo["userPrincipalName"].(string)
		if !ok {
			// Additional fallback for Microsoft consumer accounts which might not have mail or userPrincipalName
			email, ok = userInfo["id"].(string)
			if !ok {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "No identifier found in user info",
				})
			}
		}
	}

	// Check if user exists
	var user models.User
	var userId int
	err = database.DB.QueryRow("SELECT id, username, email FROM users WHERE email = ?", email).Scan(
		&userId, &user.Username, &user.Email,
	)

	if err != nil {
		// User doesn't exist, create a new one
		displayName, _ := userInfo["displayName"].(string)
		username := email
		if displayName != "" {
			// Use displayName instead of email if it's available
			username = displayName
		} else if idx := strings.Index(email, "@"); idx > 0 {
			// Otherwise use the part of the email before @
			username = email[:idx]
		}

		result, err := database.DB.Exec(
			"INSERT INTO users (username, email, created_at) VALUES (?, ?, ?)",
			username, email, time.Now(),
		)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}

		id, _ := result.LastInsertId()
		userId = int(id)
	}

	// Set user session
	sess.Set("user_id", userId)
	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save session",
		})
	}

	// Get the frontend URL from environment variable
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // Default fallback
	}

	// Check if we should use the auth callback or direct user route
	if os.Getenv("USE_AUTH_CALLBACK") == "true" {
		// Use the auth/callback route which will then redirect to user page
		redirectURL := frontendURL + "/auth/callback"
		log.Printf("Redirecting to frontend auth callback: %s", redirectURL)
		return c.Redirect(redirectURL)
	} else {
		// Redirect directly to the user view route
		redirectURL := frontendURL + "/user"
		log.Printf("Redirecting to frontend user view: %s", redirectURL)
		return c.Redirect(redirectURL)
	}
}

// Add other auth methods (Register, Login, Logout) as needed
func (ac *AuthController) Register(c *fiber.Ctx) error {
	return c.SendString("Register endpoint")
}

func (ac *AuthController) Login(c *fiber.Ctx) error {
	return c.SendString("Login endpoint")
}

func (ac *AuthController) Logout(c *fiber.Ctx) error {
	return c.SendString("Logout endpoint")
}

func (ac *AuthController) User(c *fiber.Ctx) error {
	sess, err := ac.store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get session",
		})
	}

	userID := sess.Get("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var user models.User
	err = database.DB.QueryRow("SELECT id, username, email FROM users WHERE id = ?", userID).Scan(
		&user.ID, &user.Username, &user.Email,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	return c.JSON(user)
}

// LoginPage renders the login page
func (ac *AuthController) LoginPage(c *fiber.Ctx) error {
	errorMsg := c.Query("error")

	// For API-based applications, return JSON
	if errorMsg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errorMsg,
		})
	}

	// For web applications, you could render a template here
	// If you have HTML pages, adjust accordingly
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Please log in",
	})
}
