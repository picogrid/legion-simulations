package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GetAuthorizationURLFromLegion fetches the authorization URL from the Legion API
func GetAuthorizationURLFromLegion(ctx context.Context, legionURL string) (string, error) {
	_, err := url.Parse(legionURL)
	if err != nil {
		return "", fmt.Errorf("invalid Legion URL: %w", err)
	}

	transport := &http.Transport{}
	httpClient := &http.Client{Transport: transport}

	authURLEndpoint := fmt.Sprintf("%s/v3/integrations/oauth/Cauthorization-url", legionURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURLEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get authorization URL: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Log close error but don't override the main error
			fmt.Printf("Warning: failed to close response body: %v\n", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get authorization URL: status %d", resp.StatusCode)
	}

	var authResp struct {
		AuthorizationURL string `json:"authorization_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if authResp.AuthorizationURL == "" {
		return "", fmt.Errorf("empty authorization URL in response")
	}

	return authResp.AuthorizationURL, nil
}

// GetAuthConfigFromLegion creates an AuthConfig by fetching the auth URL from Legion
func GetAuthConfigFromLegion(ctx context.Context, legionURL string) (AuthConfig, error) {
	authURL, err := GetAuthorizationURLFromLegion(ctx, legionURL)
	if err != nil {
		return AuthConfig{}, fmt.Errorf("failed to get authorization URL from Legion: %w", err)
	}

	u, err := url.Parse(authURL)
	if err != nil {
		return AuthConfig{}, fmt.Errorf("invalid authorization URL: %w", err)
	}

	// The auth URL format is typically:
	// https://auth.legion.com/auth/realms/legion/protocol/openid-connect/auth
	// We need to extract the base URL and realm

	pathParts := strings.Split(u.Path, "/")
	var keycloakURL string
	var realm string

	for i, part := range pathParts {
		if part == "realms" && i+1 < len(pathParts) {
			realm = pathParts[i+1]
			keycloakURL = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
			if strings.Contains(u.Path, "/auth/realms") {
				keycloakURL += "/auth"
			}
			break
		}
	}

	if realm == "" {
		return AuthConfig{}, fmt.Errorf("could not extract realm from authorization URL")
	}

	return AuthConfig{
		KeycloakURL: keycloakURL,
		Realm:       realm,
		ClientID:    "frontend...orion", // This is still hardcoded as it's client-specific
	}, nil
}

// AuthenticateUserWithLegion authenticates a user using the auth URL from Legion
func AuthenticateUserWithLegion(ctx context.Context, legionURL string) (*TokenManager, error) {
	config, err := GetAuthConfigFromLegion(ctx, legionURL)
	if err != nil {
		fmt.Println("⚠️  Could not fetch auth config from Legion, using defaults")
		config = DefaultAuthConfig()
	}

	return AuthenticateUser(ctx, config)
}
