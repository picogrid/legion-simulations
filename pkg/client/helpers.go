package client

import (
	"context"
	"os"
	"time"
)

// NewLegionClient creates a new Legion client with API key authentication
// This is a convenience wrapper around NewClient
func NewLegionClient(baseURL string, apiKey string) (*Legion, error) {
	cfg := Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: 30 * time.Second,
	}

	return NewClient(cfg)
}

// GetAPIKey retrieves the API key from an environment variable
func GetAPIKey(envVarName string) string {
	if envVarName == "" {
		return ""
	}
	return os.Getenv(envVarName)
}

// ValidateConnection tests the connection to Legion
func ValidateConnection(ctx context.Context, legionClient *Legion) error {
	return legionClient.ValidateConnection(ctx)
}
