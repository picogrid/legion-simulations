package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/picogrid/legion-simulations/pkg/client"
	"golang.org/x/term"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	KeycloakURL string
	Realm       string
	ClientID    string
}

// DefaultAuthConfig returns the default authentication configuration
func DefaultAuthConfig() AuthConfig {
	keycloakURL := os.Getenv("KEYCLOAK_URL")
	if keycloakURL == "" {
		keycloakURL = "https://auth.legion-staging.com"
	}

	return AuthConfig{
		KeycloakURL: keycloakURL,
		Realm:       "legion",
		ClientID:    "frontend...orion",
	}
}

// AuthenticateUser prompts for credentials and authenticates with Keycloak
func AuthenticateUser(ctx context.Context, config AuthConfig) (*TokenManager, error) {
	// Check for credentials in environment variables first
	email := os.Getenv("LEGION_EMAIL")
	password := os.Getenv("LEGION_PASSWORD")

	// If credentials are not in environment, prompt for them
	if email == "" || password == "" {
		fmt.Println("🔐 Legion Authentication")
		fmt.Println(strings.Repeat("=", 50))

		// Get username if not in environment
		if email == "" {
			fmt.Print("Email: ")
			_, err := fmt.Scanln(&email)
			if err != nil {
				return nil, err
			}
		}

		// Get password securely if not in environment
		if password == "" {
			fmt.Print("Password: ")
			passwordBytes, err := term.ReadPassword(syscall.Stdin)
			if err != nil {
				return nil, fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Println() // New line after password input
			password = string(passwordBytes)
		}
	} else {
		// Indicate we're using environment credentials
		fmt.Println("🔐 Using Legion credentials from environment")
	}

	// Create Keycloak client
	keycloakClient := NewKeycloakClient(KeycloakConfig{
		BaseURL:  config.KeycloakURL,
		Realm:    config.Realm,
		ClientID: config.ClientID,
	})

	// Authenticate
	fmt.Println("\n🔄 Authenticating...")
	tokenResp, err := keycloakClient.Authenticate(ctx, email, password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("✅ Authentication successful!")

	// Create token manager
	tokenManager := NewTokenManager(keycloakClient, tokenResp)

	return tokenManager, nil
}

// CreateAuthenticatedClient creates a Legion client with OAuth2 authentication
func CreateAuthenticatedClient(baseURL string, tokenManager *TokenManager) (*client.Legion, error) {
	return client.NewClient(client.Config{
		BaseURL:      baseURL,
		TokenManager: tokenManager,
	})
}
