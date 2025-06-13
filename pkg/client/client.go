package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/picogrid/legion-simulations/pkg/logger"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Context keys for Legion client
const (
	// OrgIDContextKey is the context key for organization ID
	OrgIDContextKey contextKey = "legion-org-id"
)

// Legion is the main client for interacting with the Legion API
type Legion struct {
	baseURL      string
	apiKey       string
	httpClient   *http.Client
	tokenManager TokenManager
}

// TokenManager interface for token management
type TokenManager interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// Config holds the configuration for the Legion client
type Config struct {
	BaseURL      string
	APIKey       string
	Timeout      time.Duration
	TokenManager TokenManager // Optional: for OAuth2 authentication
}

// NewClient creates a new Legion client with the given configuration
func NewClient(cfg Config) (*Legion, error) {
	// Parse and validate the base URL
	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Set default timeout if not provided
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Legion{
		baseURL:      u.String(),
		apiKey:       cfg.APIKey,
		tokenManager: cfg.TokenManager,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// doRequest performs an HTTP request with authentication and error handling
func (c *Legion) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	// Build the full URL
	fullURL := c.baseURL + path

	// Marshal body if provided
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set organization ID header if present in context
	if orgID, ok := ctx.Value(OrgIDContextKey).(string); ok && orgID != "" {
		req.Header.Set("X-ORG-ID", orgID)
	}

	// Set authorization header
	if c.tokenManager != nil {
		// Use OAuth2 token
		token, err := c.tokenManager.GetAccessToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	} else if c.apiKey != "" {
		// Use API key
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Perform the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed to close response body: %v", err)
			}
		}(resp.Body)
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// decodeResponse decodes a JSON response into the provided interface
func decodeResponse(resp *http.Response, v interface{}) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	if v == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// WithOrgID returns a new context with the organization ID set
func WithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, OrgIDContextKey, orgID)
}
