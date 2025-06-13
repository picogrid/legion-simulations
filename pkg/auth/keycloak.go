package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/picogrid/legion-simulations/pkg/logger"
)

// KeycloakConfig holds the configuration for Keycloak authentication
type KeycloakConfig struct {
	BaseURL  string
	Realm    string
	ClientID string
	Timeout  time.Duration
}

// TokenResponse represents the response from Keycloak token endpoint
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
}

// KeycloakClient handles authentication with Keycloak
type KeycloakClient struct {
	config     KeycloakConfig
	httpClient *http.Client
}

// NewKeycloakClient creates a new Keycloak client
func NewKeycloakClient(config KeycloakConfig) *KeycloakClient {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &KeycloakClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Authenticate performs password-based authentication
func (k *KeycloakClient) Authenticate(ctx context.Context, username, password string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", k.config.BaseURL, k.config.Realm)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", k.config.ClientID)
	data.Set("username", username)
	data.Set("password", password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("authentication request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		err := json.Unmarshal(body, &errorResp)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("authentication failed: %s", errorResp.ErrorDescription)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// RefreshToken refreshes an access token using a refresh token
func (k *KeycloakClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", k.config.BaseURL, k.config.Realm)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", k.config.ClientID)
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token refresh request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		err := json.Unmarshal(body, &errorResp)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("token refresh failed: %s", errorResp.ErrorDescription)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}
